package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

type RecommendationUsecase struct {
	pinRepo repository.PinRepository
	cache   repository.CacheRepository
}

func NewRecommendationUsecase(pinRepo repository.PinRepository, cache repository.CacheRepository) *RecommendationUsecase {
	return &RecommendationUsecase{pinRepo: pinRepo, cache: cache}
}

// ── Публичные методы ────────────────────────────────────────────────

func (u *RecommendationUsecase) GetFeed(ctx context.Context, userID int64, sessionID string, limit, offset int) ([]domain.Pin, error) {
	rankedIDs, err := u.getCachedFeed(ctx, userID)
	if err != nil || rankedIDs == nil {
		rankedIDs, err = u.buildAndCacheFeed(ctx, userID, sessionID)
		if err != nil {
			return nil, err
		}
	}

	start := offset
	end := offset + limit
	if start >= len(rankedIDs) {
		return []domain.Pin{}, nil
	}
	if end > len(rankedIDs) {
		end = len(rankedIDs)
	}

	return u.pinRepo.GetPinByIDs(ctx, rankedIDs[start:end])
}

func (u *RecommendationUsecase) GetRelated(ctx context.Context, pinID int64, limit int) ([]domain.Pin, error) {
	return u.pinRepo.FindRelated(ctx, pinID, limit)
}

// ── Сбор кандидатов ─────────────────────────────────────────────────

func (u *RecommendationUsecase) buildAndCacheFeed(ctx context.Context, userID int64, sessionID string) ([]int64, error) {
	candidates := u.collectCandidates(ctx, sessionID)
	if len(candidates) == 0 {
		return nil, nil
	}

	pinIDs := make([]int64, len(candidates))
	for i, c := range candidates {
		pinIDs[i] = c.ID
	}
	tagsMap, _ := u.pinRepo.LoadTagsForPins(ctx, pinIDs)

	ranked := rerankMMR(candidates, tagsMap, 250)
	ranked = dedupByAuthor(ranked, 2, 250) // макс 2 пина от одного автора в топ-250

	ids := make([]int64, len(ranked))
	for i, c := range ranked {
		ids[i] = c.ID
	}

	u.cacheFeed(ctx, userID, ids)
	return ids, nil
}

func (u *RecommendationUsecase) collectCandidates(ctx context.Context, sessionID string) []domain.Pin {
	var trending, related, explore []domain.Pin
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		trending, _ = u.getTrending(ctx)
	}()
	go func() {
		defer wg.Done()
		related, _ = u.pinRepo.FindTrending(ctx, 100)
	}()
	go func() {
		defer wg.Done()
		explore, _ = u.pinRepo.FindExploration(ctx, sessionID, 100)
	}()

	wg.Wait()

	seen := make(map[int64]bool, len(trending)+len(related)+len(explore))
	result := make([]domain.Pin, 0, len(trending)+len(related)+len(explore))

	addUnique := func(pins []domain.Pin) {
		for _, p := range pins {
			if !seen[p.ID] {
				seen[p.ID] = true
				result = append(result, p)
			}
		}
	}

	addUnique(trending)
	addUnique(related)
	addUnique(explore)

	return result
}

func (u *RecommendationUsecase) getTrending(ctx context.Context) ([]domain.Pin, error) {
	key := "feed:trending"
	cached, err := u.cache.Get(ctx, key)
	if err == nil && cached != nil {
		var pins []domain.Pin
		if json.Unmarshal(cached, &pins) == nil {
			return pins, nil
		}
	}

	pins, err := u.pinRepo.FindTrending(ctx, 200)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(pins)
	_ = u.cache.Set(ctx, key, data, 2*time.Minute)
	return pins, nil
}

// ═══════════════════════════════════════════════════════════════════════
// MMR (Maximum Marginal Relevance)
//
// Выбираем пины так, чтобы они были и релевантными, и разнообразными.
//
// MMR = λ * relevance(i) - (1-λ) * max(similarity(i, selected))
//
// similarity считаем через Jaccard Index на тегах: |A∩B| / |A∪B|
// Если тегов нет — similarity = 0 (нет штрафа).
// ═══════════════════════════════════════════════════════════════════════

func rerankMMR(candidates []domain.Pin, tagsMap map[int64][]string, topN int) []domain.Pin {
	if len(candidates) <= topN {
		return candidates
	}

	// relevance = created_at unix (свежие выше)
	scored := make([]domain.Pin, len(candidates))
	copy(scored, candidates)

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].CreatedAt.Unix() > scored[j].CreatedAt.Unix()
	})

	selected := make([]domain.Pin, 0, topN)
	remaining := scored

	for len(selected) < topN && len(remaining) > 0 {
		bestIdx := 0
		bestMMR := -1.0

		for i, cand := range remaining {
			relevance := float64(cand.CreatedAt.Unix())
			penalty := maxTagOverlap(cand.ID, tagsMap, selected, tagsMap)
			mmr := 0.5*relevance - 0.5*penalty

			if mmr > bestMMR {
				bestMMR = mmr
				bestIdx = i
			}
		}

		selected = append(selected, remaining[bestIdx])
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	return selected
}

// maxTagOverlap — максимальный Jaccard Index между пинами и любым из selected.
func maxTagOverlap(pinID int64, tagsMap map[int64][]string, selected []domain.Pin, selectedTags map[int64][]string) float64 {
	tagsA := tagsMap[pinID]
	if len(tagsA) == 0 {
		return 0
	}

	maxSim := 0.0
	for _, sel := range selected {
		tagsB := selectedTags[sel.ID]
		if len(tagsB) == 0 {
			continue
		}

		setB := make(map[string]bool, len(tagsB))
		for _, t := range tagsB {
			setB[t] = true
		}

		intersection := 0
		for _, t := range tagsA {
			if setB[t] {
				intersection++
			}
		}

		union := len(tagsA) + len(tagsB) - intersection
		if union == 0 {
			continue
		}

		sim := float64(intersection) / float64(union)
		if sim > maxSim {
			maxSim = sim
		}
	}

	return maxSim
}

// dedupByAuthor — не даёт одному автору занять всю ленту.
// После MMR проходим по отсортированному списку и пропускаем
// пины авторов, которые уже набрали maxPerAuthor.
func dedupByAuthor(pins []domain.Pin, maxPerAuthor, maxTotal int) []domain.Pin {
	authorCount := make(map[int64]int)
	result := make([]domain.Pin, 0, maxTotal)

	for _, p := range pins {
		if len(result) >= maxTotal {
			break
		}
		if authorCount[p.AuthorId] >= maxPerAuthor {
			continue
		}
		authorCount[p.AuthorId]++
		result = append(result, p)
	}

	return result
}

// ── Кеш готовой ленты ───────────────────────────────────────────────

func feedKey(userID int64) string { return fmt.Sprintf("feed:ranked:%d", userID) }

func (u *RecommendationUsecase) getCachedFeed(ctx context.Context, userID int64) ([]int64, error) {
	data, err := u.cache.Get(ctx, feedKey(userID))
	if err != nil || data == nil {
		return nil, err
	}
	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (u *RecommendationUsecase) cacheFeed(ctx context.Context, userID int64, ids []int64) {
	data, _ := json.Marshal(ids)
	_ = u.cache.Set(ctx, feedKey(userID), data, 5*time.Minute)
}

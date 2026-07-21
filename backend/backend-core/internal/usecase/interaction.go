package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

// InteractionUsecase — бизнес-логика сбора событий.
//
// Ключевая фича: batch-буфер. Вместо того чтобы писать каждое событие
// отдельным INSERT (что убило бы БД под нагрузкой), мы накапливаем
// события в памяти и сбрасываем в БД пачкой раз в N секунд или
// после N событий.
//
// Это trade-off:
//   (-) При краше сервера теряем до batchSize событий (некритично для аналитики)
//   (+) В 100 раз меньше нагрузка на PostgreSQL
type InteractionUsecase struct {
	repo    repository.InteractionRepository
	buffer  *interactionBuffer
}

// NewInteractionUsecase создаёт usecase с запущенным буфером.
// Важно: вызовите Close() при завершении работы сервера, чтобы
// сбросить остаток буфера в БД.
func NewInteractionUsecase(repo repository.InteractionRepository, batchSize int, flushInterval time.Duration) *InteractionUsecase {
	uc := &InteractionUsecase{
		repo:   repo,
		buffer: newInteractionBuffer(repo, batchSize, flushInterval),
	}
	uc.buffer.start()
	return uc
}

// Save принимает событие от клиента, валидирует и кладёт в буфер.
func (u *InteractionUsecase) Save(ctx context.Context, userID *int64, sessionID string, pinID int64, eventType string, ts time.Time, meta map[string]interface{}) error {
	// ── Валидация ────────────────────────────────────────────────

	if sessionID == "" {
		return errors.New("session_id is required")
	}
	if pinID <= 0 {
		return errors.New("pin_id must be positive")
	}
	if !domain.AllowedEventTypes[eventType] {
		return fmt.Errorf("unsupported event type: %q", eventType)
	}

	// ── Сериализуем meta в JSON ──────────────────────────────────
	var metaStr string
	if len(meta) > 0 {
		data, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshal meta: %w", err)
		}
		metaStr = string(data)
	}

	event := domain.Interaction{
		UserID:    userID,
		SessionID: sessionID,
		PinID:     pinID,
		EventType: eventType,
		TS:        ts,
		MetaJSON:  metaStr,
	}

	u.buffer.add(event)
	return nil
}

// Close сбрасывает остаток буфера в БД.
func (u *InteractionUsecase) Close() error {
	u.buffer.stop()
	return nil
}

// ═══════════════════════════════════════════════════════════════════════
// Batch-буфер
// ═══════════════════════════════════════════════════════════════════════

type interactionBuffer struct {
	repo          repository.InteractionRepository
	batchSize     int
	flushInterval time.Duration

	mu      sync.Mutex
	buffer  []domain.Interaction
	done    chan struct{}
	stopped bool
}

func newInteractionBuffer(repo repository.InteractionRepository, batchSize int, flushInterval time.Duration) *interactionBuffer {
	return &interactionBuffer{
		repo:          repo,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]domain.Interaction, 0, batchSize),
		done:          make(chan struct{}),
	}
}

func (b *interactionBuffer) start() {
	go func() {
		ticker := time.NewTicker(b.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Каждые N секунд сбрасываем буфер
				b.flush()
			case <-b.done:
				// Сервер завершает работу — сбрасываем остаток
				b.flush()
				return
			}
		}
	}()
}

func (b *interactionBuffer) stop() {
	b.mu.Lock()
	b.stopped = true
	b.mu.Unlock()
	close(b.done)
}

func (b *interactionBuffer) add(event domain.Interaction) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, event)

	// Если буфер переполнен — сбрасываем немедленно
	if len(b.buffer) >= b.batchSize {
		b.flushLocked()
	}
}

func (b *interactionBuffer) flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}

func (b *interactionBuffer) flushLocked() {
	if len(b.buffer) == 0 {
		return
	}

	batch := b.buffer
	b.buffer = make([]domain.Interaction, 0, b.batchSize)

	// Не блокируем ответ клиента — сбрасываем асинхронно
	go func() {
		if err := b.repo.BatchSave(context.Background(), batch); err != nil {
			log.Printf("flush interactions: %v (dropped %d events)", err, len(batch))
		}
	}()
}

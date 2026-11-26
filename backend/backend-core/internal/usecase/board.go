package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

type InteractionRecorder interface {
	Save(ctx context.Context, userId *int64, sessionID string, pinID int64, eventType string, ts time.Time, meta map[string]interface{}) error
}

// BoardUsecase — бизнес-логика работы с досками.
//
// Здесь проверяются правила, которые не относятся к чистоте данных (это задача repository):
//   - Кто может создавать доски? Любой авторизованный пользователь.
//   - Кто может сохранять пин в доску? Только владелец доски.
//   - Можно ли сохранить один пин дважды? Нет — usecase проверяет дубликат.
//   - Можно ли удалить чужой пин из чужой доски? Нет — проверка владельца.
type BoardUsecase struct {
	Repo    repository.BoardRepository
	PinRepo interface {
		GetPinByID(ctx context.Context, id int64) (*domain.Pin, error)
		EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error)
	}
	Interactions InteractionRecorder
}



const defaultBoardTitle = "Избранное"

// NewBoardUsecase создаёт BoardUsecase.
// PinRepo передаётся как интерфейс с одним методом GetPinByID,
// чтобы не тащить весь PinRepository.
func NewBoardUsecase(repo repository.BoardRepository, pinRepo interface {
	GetPinByID(ctx context.Context, id int64) (*domain.Pin, error)
	EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error)
}, interactions InteractionRecorder) *BoardUsecase {
	return &BoardUsecase{
		Repo:         repo,
		PinRepo:      pinRepo,
		Interactions: interactions,
	}
}

// ── Создание доски ───────────────────────────────────────────────────

func (u *BoardUsecase) Create(ctx context.Context, ownerID int64, title string, isPrivate bool) (*domain.Board, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("title is required")
	}
	if len(title) > 100 {
		return nil, errors.New("title is too long (max 100 characters)")
	}

	board := &domain.Board{
		OwnerID:   ownerID,
		Title:     title,
		IsPrivate: isPrivate,
	}

	if err := u.Repo.CreateBoard(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

// ── Список досок пользователя ────────────────────────────────────────

func (u *BoardUsecase) ListUserBoards(ctx context.Context, userID int64) ([]domain.BoardWithRecentPins, error) {
	return u.Repo.ListUserBoards(ctx, userID)
}

// ── Получение одной доски ─────────────────────────────────────────────

func (u *BoardUsecase) GetBoardByID(ctx context.Context, boardID int64) (*domain.Board, error) {
	return u.Repo.GetBoardByID(ctx, boardID)
}

func (u *BoardUsecase) RemoveBoardByID(ctx context.Context, boardID int64, userID int64) error {
	board, err := u.Repo.GetBoardByID(ctx, boardID)
	if err != nil {
		return errors.New("board not found")
	}
	if board.OwnerID != userID {
		return errors.New("you don't own this board")
	}

	err = u.Repo.RemoveBoardByID(ctx, boardID, userID)

	if err != nil {
		return err
	}
	return nil
}

// ── Сохранение пина в доску ──────────────────────────────────────────

// SavePinToBoard сохраняет пин в доску.
// Проверяет:
//   - Пин существует.
//   - Доска существует и принадлежит пользователю.
//   - Пин ещё не сохранён в эту доску (избегаем дубликата).
func (u *BoardUsecase) SavePinToBoard(ctx context.Context, userID, boardID, pinID int64) (*domain.BoardPin, error) {
	// Проверяем, что доска существует и принадлежит этому пользователю
	board, err := u.Repo.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, errors.New("board not found")
	}
	if board.OwnerID != userID {
		return nil, errors.New("you don't own this board")
	}

	// Проверяем, что пин существует (не даём сохранять несуществующие пины)
	if _, err := u.PinRepo.GetPinByID(ctx, pinID); err != nil {
		return nil, errors.New("pin not found")
	}

	// Сохраняем. Если пин уже есть — БД вернёт ошибку нарушая PRIMARY KEY,
	// и SavePinToBoard внутри repository вернёт ошибку.
	bp, err := u.Repo.SavePinToBoard(ctx, boardID, pinID)
	if err != nil {
		// Конвертируем сырую ошибку БД в понятное сообщение
		return nil, errors.New("pin is already saved or save failed")
	}

	return bp, nil
}

// GetDefaultBoard возвращает доску по умолчанию для пользователя
func (u *BoardUsecase) GetDefaultBoard(ctx context.Context, userID int64) (*domain.Board, error) {
	board, err := u.Repo.GetDefaultBoard(ctx, userID)

	if err != nil {
		return nil, err
	}

	return board, nil


//
// 	newBoard := &domain.Board{
// 		OwnerID: userID,
// 		Title:   defaultBoardTitle,
// 		IsDefault: true,
// 	}
// 	if err := u.Repo.CreateBoard(ctx, newBoard); err != nil {
// 		board, err := u.Repo.GetDefaultBoard(ctx, userID) //Не оптимиальное решение но для mvp пойдет)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return board, nil
// 	};
//
// 	return nil, err
}

// ── Удаление пина из доски ───────────────────────────────────────────

func (u *BoardUsecase) RemovePinFromBoard(ctx context.Context, userID, boardID, pinID int64) error {
	board, err := u.Repo.GetBoardByID(ctx, boardID)
	if err != nil {
		return errors.New("board not found")
	}
	if board.OwnerID != userID {
		return errors.New("you don't own this board")
	}

	if err := u.Repo.RemovePinFromBoard(ctx, boardID, pinID); err != nil {
		return err
	}

	return nil
}

func (u *BoardUsecase) SavePinToProfile(ctx context.Context, userID, pinID int64, sessionId string) error {
	board, err := u.GetDefaultBoard(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := u.PinRepo.GetPinByID(ctx, pinID); err != nil {
		return err
	}

	if _, err := u.Repo.SavePinToBoard(ctx, board.ID, pinID); err != nil {
		return err
	}
	u.recordInteraction(ctx, &userID, sessionId, "save", pinID, map[string]interface{}{"board_id": board.ID})
	return nil
}


func (u *BoardUsecase) RemovePinFromProfile(ctx context.Context, userID, pinID int64, sessionId string) error {
	board, err := u.GetDefaultBoard(ctx, userID)
	if err != nil {
		return nil
	}

	if err := u.Repo.RemovePinFromBoard(ctx, board.ID, pinID); err != nil {
		return err
	}
	u.recordInteraction(ctx, &userID, sessionId, "unsave", pinID, map[string]interface{}{"board_id": board.ID})
	return nil
}


// ── Получение пинов доски ─────────────────────────────────────────────

func (u *BoardUsecase) GetBoardPins(ctx context.Context, boardID int64) ([]domain.Pin, error) {
	return u.Repo.GetBoardPins(ctx, boardID)
}

func (u *BoardUsecase) EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error) {
	return u.PinRepo.EnrichPins(ctx, pins, viewerID)
}


func (u *BoardUsecase) recordInteraction(ctx context.Context, userID *int64, sessionID, eventType string, pinID int64, meta map[string]interface{}) {
	if u.Interactions == nil || sessionID == "" {
		return
	}
	if err := u.Interactions.Save(ctx, userID, sessionID, pinID, eventType, time.Now().UTC(), meta); err != nil {
		log.Printf("record interaction %s: %v", eventType, err)
	}
}

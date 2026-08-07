package usecase

import (
	"context"
	"errors"
	"strings"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

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
}

// NewBoardUsecase создаёт BoardUsecase.
// PinRepo передаётся как интерфейс с одним методом GetPinByID,
// чтобы не тащить весь PinRepository.
func NewBoardUsecase(repo repository.BoardRepository, pinRepo interface {
	GetPinByID(ctx context.Context, id int64) (*domain.Pin, error)
	EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error)
}) *BoardUsecase {
	return &BoardUsecase{
		Repo:    repo,
		PinRepo: pinRepo,
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

func (u *BoardUsecase) ListUserBoards(ctx context.Context, userID int64) ([]domain.Board, error) {
	return u.Repo.ListUserBoards(ctx, userID)
}

// ── Получение одной доски ─────────────────────────────────────────────

func (u *BoardUsecase) GetBoardByID(ctx context.Context, boardID int64) (*domain.Board, error) {
	return u.Repo.GetBoardByID(ctx, boardID)
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

// ── Получение пинов доски ─────────────────────────────────────────────

func (u *BoardUsecase) GetBoardPins(ctx context.Context, boardID int64) ([]domain.Pin, error) {
	return u.Repo.GetBoardPins(ctx, boardID)
}

func (u *BoardUsecase) EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error) {
	return u.PinRepo.EnrichPins(ctx, pins, viewerID)
}

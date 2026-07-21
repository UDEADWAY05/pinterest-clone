package repository

import (
	"context"
	"pinterest-core/internal/domain"
)


type BoardRepository interface {
	CreateBoard(ctx context.Context, board *domain.Board) error
	ListUserBoards(ctx context.Context, userID int64) ([]domain.Board, error)
	GetBoardByID(ctx context.Context, boardID int64) (*domain.Board, error)
	SavePinToBoard(ctx context.Context, boardID, pinID int64) (*domain.BoardPin, error)
	RemovePinFromBoard(ctx context.Context, boardID, pinID int64) error
	GetBoardPins(ctx context.Context, boardID int64) ([]domain.Pin, error)
}

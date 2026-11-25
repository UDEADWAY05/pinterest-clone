package repository

import (
	"context"
	"pinterest-core/internal/domain"
)


type BoardRepository interface {
	CreateBoard(ctx context.Context, board *domain.Board) error
	ListUserBoards(ctx context.Context, userID int64) ([]domain.BoardWithRecentPins, error)
	GetBoardByID(ctx context.Context, boardID int64) (*domain.Board, error)
	RemoveBoardByID(ctx context.Context, boardID int64, userID int64) error
	GetDefaultBoard(ctx context.Context, userID int64) (*domain.Board, error) //TODO: переделать, сохранять как обычный борд, заметил на этапе коммита
	SavePinToBoard(ctx context.Context, boardID, pinID int64) (*domain.BoardPin, error)
	RemovePinFromBoard(ctx context.Context, boardID, pinID int64) error
	GetBoardPins(ctx context.Context, boardID int64) ([]domain.Pin, error)
}

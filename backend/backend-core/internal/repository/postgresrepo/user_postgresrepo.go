package postgresrepo

import (
	"context"
	"fmt"
	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	Pool *pgxpool.Pool
}



func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{Pool: pool}
}


func (r *UserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, created_at
		FROM users
		ORDER BY created_at DESC`)

	if err != nil {
		fmt.Print(err)
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.CreatedAt); err != nil {
			fmt.Print(err)
			return nil, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User

	err := r.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, created_at
		FROM users WHERE id = $1`,
		id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, mapNoRows(err, repository.ErrUserNotFound)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, created_at
		FROM users
		WHERE email = $1`,
		email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, mapNoRows(err, repository.ErrUserNotFound)
	}

	return &user, nil
}

func mapNoRows(err error, notFound error) error {
	if err.Error() == "no rows in result set" {
		return notFound
	}
	return fmt.Errorf("database query failed: %w", err)
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		user.Email, user.PasswordHash, user.DisplayName, user.AvatarURL,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

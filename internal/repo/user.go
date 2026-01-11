package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kkonst40/isso/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func New(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) GetAll(ctx context.Context) ([]model.User, error) {
	const query = `
		SELECT *
		FROM users
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("internal db error")
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.PasswordHash,
			&user.TokenID,
		); err != nil {
			return nil, fmt.Errorf("internal db error")
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("internal db error")
	}

	return users, nil
}

func (r *UserRepo) GetByID(ctx context.Context, ID uuid.UUID) (*model.User, error) {
	const query = `
		SELECT *
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, ID).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user (%v) not found", ID)
	}
	if err != nil {
		return nil, fmt.Errorf("internal db error")
	}

	return &user, nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	const query = `
		SELECT id, login, password_hash, token_id
		FROM users
		WHERE login = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.TokenID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user (%v) not found", login)
	}
	if err != nil {
		return nil, fmt.Errorf("internal db error: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO users (id, login, password_hash, token_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.TokenID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// unique violation
			// error 23505 -> Возвращать 409 Conflict
			if pgErr.Code == "23505" {
				return fmt.Errorf("user with login (%s) already exists", user.Login)
			}
		}

		return fmt.Errorf("internal db error: %w", err)
	}

	return nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	const query = `
		UPDATE users
		SET 
			login = $1,
			password_hash = $2,
			token_id = $3
		WHERE id = $4
	`

	res, err := r.db.ExecContext(ctx, query, user.Login, user.Login, user.TokenID, user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// unique violation
			if pgErr.Code == "23505" {
				return fmt.Errorf("cannot update: login (%s) is already taken", user.Login)
			}
		}
		return fmt.Errorf("internal db error: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("internal db error")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user (%v) not found", user.ID)
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, ID uuid.UUID) error {
	const query = `
		DELETE FROM users
		WHERE id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, ID); err != nil {
		return fmt.Errorf("internal db error")
	}

	return nil
}

func (r *UserRepo) Exist(ctx context.Context, ID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, ID).Scan(
		&exists,
	)

	if err != nil {
		return false, fmt.Errorf("internal db error")
	}

	return exists, nil
}

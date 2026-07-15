package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	errs "github.com/kkonst40/sso-service/internal/domain/errors"
	"github.com/kkonst40/sso-service/internal/domain/model"
	"github.com/kkonst40/sso-service/internal/repo"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) GetByID(ctx context.Context, ID uuid.UUID) (model.User, error) {
	const query = `
		SELECT id, login, password_hash
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
		return model.User{}, errs.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return user, nil
}

func (r *Repo) GetByLogin(ctx context.Context, login string) (model.User, error) {
	const query = `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err == sql.ErrNoRows {
		return model.User{}, errs.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return user, nil
}

func (r *Repo) GetLoginsByIDs(ctx context.Context, IDs []uuid.UUID) ([]model.UserInfo, error) {
	if len(IDs) == 0 {
		return []model.UserInfo{}, nil
	}

	const query = `
		SELECT id, login
        FROM users
        WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, IDs)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer rows.Close()

	var users []model.UserInfo
	for rows.Next() {
		var u model.UserInfo
		if err := rows.Scan(&u.ID, &u.Login); err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repo) GetIDsByLogins(ctx context.Context, logins []string) ([]model.UserInfo, error) {
	if len(logins) == 0 {
		return []model.UserInfo{}, nil
	}

	const query = `
		SELECT id, login
        FROM users
        WHERE login = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, logins)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer rows.Close()

	var users []model.UserInfo
	for rows.Next() {
		var u model.UserInfo
		if err := rows.Scan(&u.ID, &u.Login); err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repo) Create(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO users (id, login, password_hash)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Login,
		user.PasswordHash,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == repo.UniqueViolationCode {
				if pgErr.ConstraintName == "users_login" {
					return errs.ErrLoginExists
				}
			}
		}

		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *Repo) Update(ctx context.Context, user *model.User) error {
	const query = `
		UPDATE users
		SET 
			login = $1,
			password_hash = $2
		WHERE id = $3
	`

	res, err := r.db.ExecContext(ctx, query, user.Login, user.PasswordHash, user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == repo.UniqueViolationCode {
				if pgErr.ConstraintName == "users_login_key" {
					return fmt.Errorf("%w: login '%s' taken", errs.ErrLoginExists, user.Login)
				}
			}
		}
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return errs.ErrUserNotFound
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, ID uuid.UUID) error {
	const query = `
		DELETE FROM users
		WHERE id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, ID); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *Repo) Exist(ctx context.Context, IDs []uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT id
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, IDs)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer rows.Close()

	existIDs := []uuid.UUID{}
	for rows.Next() {
		var ID uuid.UUID
		if err := rows.Scan(&ID); err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
		}

		existIDs = append(existIDs, ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return existIDs, nil
}

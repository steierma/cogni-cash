package postgres

import (
	"context"
	"errors"
	"log/slog"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *slog.Logger) *UserRepository {
	return &UserRepository{pool: pool, logger: logger}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	query := `SELECT id, username, password_hash, email, full_name, address, role FROM users WHERE username = $1`
	var user entity.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName, &user.Address, &user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, errors.New("user not found")
		}
		return entity.User{}, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	query := `SELECT id, username, password_hash, email, full_name, address, role FROM users WHERE id = $1`
	var user entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName, &user.Address, &user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, errors.New("user not found")
		}
		return entity.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, "SELECT id FROM users WHERE username = 'admin'").Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, errors.New("admin user not found")
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (r *UserRepository) FindAll(ctx context.Context, search string) ([]entity.User, error) {
	var rows pgx.Rows
	var err error

	if search != "" {
		query := `
			SELECT id, username, email, full_name, address, role 
			FROM users 
			WHERE username ILIKE $1 OR email ILIKE $1 OR full_name ILIKE $1
			ORDER BY username ASC`
		rows, err = r.pool.Query(ctx, query, "%"+search+"%")
	} else {
		query := `SELECT id, username, email, full_name, address, role FROM users ORDER BY username ASC`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.Address, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, email, full_name, address, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.Email, user.FullName, user.Address, user.Role,
	)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user entity.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, full_name = $3, address = $4, role = $5
		WHERE id = $6`
	cmdTag, err := r.pool.Exec(ctx, query,
		user.Username, user.Email, user.FullName, user.Address, user.Role, user.ID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) Upsert(ctx context.Context, user entity.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, email, full_name, address, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (username)
		DO UPDATE SET
			email = EXCLUDED.email,
			full_name = EXCLUDED.full_name,
			address = EXCLUDED.address,
			role = EXCLUDED.role`
	// NOTE: password_hash is intentionally excluded from the UPDATE clause.
	// The password set on first insert is only ever changed through the
	// dedicated UpdatePassword method (Settings page / resetpw CLI).
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.Email, user.FullName, user.Address, user.Role,
	)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, newHash, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

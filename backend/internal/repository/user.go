package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrNicknameExists    = errors.New("nickname already exists")
	ErrVersionConflict   = errors.New("version conflict: another admin is modifying this user")
	ErrInvalidRole       = errors.New("invalid role")
	ErrInvalidPermission = errors.New("invalid permission")
)

// UserRepository handles user database operations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, nickname, email_verify_token, role, permissions)
		VALUES ($1, $2, $3, $4, 'USER', '[]'::jsonb)
		RETURNING id, role, permissions, version, created_at, updated_at
	`

	var permissionsJSON []byte
	err := r.db.Pool.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Nickname,
		user.EmailVerifyToken,
	).Scan(&user.ID, &user.Role, &permissionsJSON, &user.Version, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check for unique constraint violations
		if strings.Contains(err.Error(), "users_email_key") {
			return ErrEmailExists
		}
		if strings.Contains(err.Error(), "users_nickname_key") {
			return ErrNicknameExists
		}
		return err
	}

	// Parse permissions
	if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
		user.Permissions = []string{}
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, nickname, role, permissions, version, email_verified,
		       email_verify_token, refresh_token, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	var permissionsJSON []byte
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nickname,
		&user.Role,
		&permissionsJSON,
		&user.Version,
		&user.EmailVerified,
		&user.EmailVerifyToken,
		&user.RefreshToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Parse permissions
	if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
		user.Permissions = []string{}
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, nickname, role, permissions, version, email_verified,
		       email_verify_token, refresh_token, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &model.User{}
	var permissionsJSON []byte
	err := r.db.Pool.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nickname,
		&user.Role,
		&permissionsJSON,
		&user.Version,
		&user.EmailVerified,
		&user.EmailVerifyToken,
		&user.RefreshToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Parse permissions
	if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
		user.Permissions = []string{}
	}

	return user, nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.Pool.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ExistsByNickname checks if a user with the given nickname exists
func (r *UserRepository) ExistsByNickname(ctx context.Context, nickname string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE nickname = $1)`

	var exists bool
	err := r.db.Pool.QueryRowContext(ctx, query, nickname).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateRefreshToken updates the user's refresh token
func (r *UserRepository) UpdateRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	query := `UPDATE users SET refresh_token = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.db.Pool.ExecContext(ctx, query, token, userID)
	return err
}

// ClearRefreshToken clears the user's refresh token
func (r *UserRepository) ClearRefreshToken(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET refresh_token = NULL, updated_at = NOW() WHERE id = $1`

	_, err := r.db.Pool.ExecContext(ctx, query, userID)
	return err
}

// SetPasswordResetToken sets the password reset token and expiry
func (r *UserRepository) SetPasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expires time.Time) error {
	query := `UPDATE users SET password_reset_token = $1, password_reset_expires = $2, updated_at = NOW() WHERE id = $3`

	_, err := r.db.Pool.ExecContext(ctx, query, token, expires, userID)
	return err
}

// GetByPasswordResetToken retrieves a user by password reset token
func (r *UserRepository) GetByPasswordResetToken(ctx context.Context, token string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, nickname, role, email_verified,
		       password_reset_token, password_reset_expires, created_at, updated_at
		FROM users
		WHERE password_reset_token = $1
	`

	user := &model.User{}
	err := r.db.Pool.QueryRowContext(ctx, query, token).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nickname,
		&user.Role,
		&user.EmailVerified,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// UpdatePassword updates the user's password and clears reset token
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, password_reset_token = NULL, password_reset_expires = NULL, updated_at = NOW() WHERE id = $2`

	_, err := r.db.Pool.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// ListUsers retrieves a paginated list of users with optional role filter
func (r *UserRepository) ListUsers(ctx context.Context, page, limit int, search, roleFilter string) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if search != "" {
		whereClause += " AND (email ILIKE '%' || $" + string(rune('0'+argIdx)) + " || '%' OR nickname ILIKE '%' || $" + string(rune('0'+argIdx)) + " || '%')"
		args = append(args, search)
		argIdx++
	}

	if roleFilter != "" {
		whereClause += " AND role = $" + string(rune('0'+argIdx))
		args = append(args, roleFilter)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users
	args = append(args, limit, offset)
	query := `
		SELECT id, email, nickname, role, permissions, version, email_verified, created_at, updated_at
		FROM users
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))

	rows, err := r.db.Pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		var permissionsJSON []byte
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Nickname,
			&user.Role,
			&permissionsJSON,
			&user.Version,
			&user.EmailVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		// Parse permissions
		if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
			user.Permissions = []string{}
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListUsersSimple retrieves a paginated list of users (simplified query)
func (r *UserRepository) ListUsersSimple(ctx context.Context, page, limit int, search string) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM users WHERE ($1 = '' OR email ILIKE '%' || $1 || '%' OR nickname ILIKE '%' || $1 || '%')`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users
	query := `
		SELECT id, email, nickname, role, permissions, version, email_verified, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR email ILIKE '%' || $1 || '%' OR nickname ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		var permissionsJSON []byte
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Nickname,
			&user.Role,
			&permissionsJSON,
			&user.Version,
			&user.EmailVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		// Parse permissions
		if err := json.Unmarshal(permissionsJSON, &user.Permissions); err != nil {
			user.Permissions = []string{}
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUserRole updates the user's role with optimistic locking
func (r *UserRepository) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string, expectedVersion int) (int, error) {
	query := `
		UPDATE users
		SET role = $1, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND version = $3
		RETURNING version
	`

	var newVersion int
	err := r.db.Pool.QueryRowContext(ctx, query, role, userID, expectedVersion).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrVersionConflict
		}
		return 0, err
	}

	return newVersion, nil
}

// UpdateUserPermissions updates the user's permissions with optimistic locking
func (r *UserRepository) UpdateUserPermissions(ctx context.Context, userID uuid.UUID, permissions []string, expectedVersion int) (int, error) {
	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return 0, err
	}

	query := `
		UPDATE users
		SET permissions = $1::jsonb, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND version = $3
		RETURNING version
	`

	var newVersion int
	err = r.db.Pool.QueryRowContext(ctx, query, string(permissionsJSON), userID, expectedVersion).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrVersionConflict
		}
		return 0, err
	}

	return newVersion, nil
}

// CountUsers returns the total number of users
func (r *UserRepository) CountUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.Pool.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// CountUsersByRole returns the count of users by role
func (r *UserRepository) CountUsersByRole(ctx context.Context) (map[string]int, error) {
	query := `SELECT role, COUNT(*) FROM users GROUP BY role`
	rows, err := r.db.Pool.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		result[role] = count
	}

	return result, rows.Err()
}

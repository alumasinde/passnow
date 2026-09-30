package platform

import (
    "context"
    "database/sql"
    "errors"
)

var ErrAdminNotFound = errors.New("platform: admin not found")

type Admin struct {
    UserID int64
    Role string
    Status string
}

type AdminRepository struct { db *sql.DB }

func NewAdminRepository(db *sql.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) ByUserID(ctx context.Context, userID int64) (*Admin, error) {
    a := &Admin{}
    err := r.db.QueryRowContext(ctx,
        "SELECT user_id, role, status FROM platform_admins WHERE user_id = ? LIMIT 1", userID,
    ).Scan(&a.UserID, &a.Role, &a.Status)
    if errors.Is(err, sql.ErrNoRows) { return nil, ErrAdminNotFound }
    if err != nil { return nil, err }
    return a, nil
}

func (r *AdminRepository) Grant(ctx context.Context, userID int64, role string) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO platform_admins (user_id, role, status)
        VALUES (?, ?, 'active')
        ON DUPLICATE KEY UPDATE role = VALUES(role), status = 'active', updated_at = NOW()`, userID, role)
    return err
}

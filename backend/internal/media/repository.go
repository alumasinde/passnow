package media

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("media: not found")

type File struct {
	ID          int64     `json:"id"`
	PublicID    string    `json:"public_id"`
	OriginalName string   `json:"original_name"`
	StoragePath string    `json:"-"`
	MimeType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Purpose     string    `json:"purpose"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, f *File) (int64, error) {
	res, err := r.db.ExecContext(ctx,`
		INSERT INTO media_files
			(public_id, original_name, storage_path, mime_type, size_bytes, purpose, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())`,
		f.PublicID, f.OriginalName, f.StoragePath, f.MimeType, f.SizeBytes, f.Purpose, f.CreatedBy,
	)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

func scanFile(s interface{ Scan(...any) error }) (*File, error) {
	var f File
	if err := s.Scan(&f.ID, &f.PublicID, &f.OriginalName, &f.StoragePath, &f.MimeType, &f.SizeBytes, &f.Purpose, &f.CreatedBy, &f.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, ErrNotFound }
		return nil, err
	}
	return &f, nil
}

const fileColumns = `id, public_id, original_name, storage_path, mime_type, size_bytes, purpose, created_by, created_at`

func (r *Repository) ByPublicID(ctx context.Context, publicID string) (*File, error) {
	return scanFile(r.db.QueryRowContext(ctx, `SELECT `+fileColumns+` FROM media_files WHERE public_id = ? LIMIT 1`, publicID))
}

func (r *Repository) List(ctx context.Context, limit int) ([]File, error) {
	if limit <= 0 || limit > 200 { limit = 100 }
	rows, err := r.db.QueryContext(ctx, `SELECT `+fileColumns+` FROM media_files ORDER BY id DESC LIMIT ?`, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	out := make([]File, 0)
	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil { return nil, err }
		out = append(out, *f)
	}
	return out, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id int64) (*File, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return nil, err }
	defer tx.Rollback()
	f, err := scanFile(tx.QueryRowContext(ctx, `SELECT `+fileColumns+` FROM media_files WHERE id = ? FOR UPDATE`, id))
	if err != nil { return nil, err }
	if _, err := tx.ExecContext(ctx, `DELETE FROM media_files WHERE id = ?`, id); err != nil { return nil, err }
	if err := tx.Commit(); err != nil { return nil, err }
	return f, nil
}

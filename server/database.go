package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database wraps the PostgreSQL connection pool
type Database struct {
	pool *pgxpool.Pool
}

// NewDatabase creates a new database connection
func NewDatabase(databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Connection pool settings optimized for Cloud Run
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{pool: pool}

	// Run migrations
	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	db.pool.Close()
}

// Migrate runs database migrations
func (db *Database) Migrate() error {
	ctx := context.Background()

	migrations := []string{
		// Users table - synced from Clerk
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			clerk_id TEXT UNIQUE NOT NULL,
			email TEXT,
			name TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		// Photos table
		`CREATE TABLE IF NOT EXISTS photos (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			filename TEXT NOT NULL,
			original_filename TEXT NOT NULL,
			gcs_path TEXT NOT NULL,
			thumb_gcs_path TEXT,
			size_bytes BIGINT NOT NULL,
			content_type TEXT NOT NULL,
			width INT,
			height INT,
			taken_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,

		// Memory books table
		`CREATE TABLE IF NOT EXISTS books (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT,
			cover_photo_id TEXT REFERENCES photos(id),
			status TEXT DEFAULT 'draft',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		// Book pages table
		`CREATE TABLE IF NOT EXISTS pages (
			id TEXT PRIMARY KEY,
			book_id TEXT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
			page_number INT NOT NULL,
			title TEXT,
			description TEXT,
			theme TEXT,
			background_gcs_path TEXT,
			layout_json JSONB,
			status TEXT DEFAULT 'draft',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(book_id, page_number)
		)`,

		// Page photos junction table
		`CREATE TABLE IF NOT EXISTS page_photos (
			page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
			photo_id TEXT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
			position INT DEFAULT 0,
			PRIMARY KEY (page_id, photo_id)
		)`,

		// Photo clusters (for AI grouping)
		`CREATE TABLE IF NOT EXISTS clusters (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT,
			description TEXT,
			theme TEXT,
			date TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		// Cluster photos junction table
		`CREATE TABLE IF NOT EXISTS cluster_photos (
			cluster_id TEXT NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
			photo_id TEXT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
			PRIMARY KEY (cluster_id, photo_id)
		)`,

		// Page drafts table
		`CREATE TABLE IF NOT EXISTS page_drafts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			cluster_id TEXT REFERENCES clusters(id) ON DELETE SET NULL,
			title TEXT,
			description TEXT,
			theme TEXT,
			background_gcs_path TEXT,
			status TEXT DEFAULT 'draft',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,

		// Draft photos junction table
		`CREATE TABLE IF NOT EXISTS draft_photos (
			draft_id TEXT NOT NULL REFERENCES page_drafts(id) ON DELETE CASCADE,
			photo_id TEXT NOT NULL REFERENCES photos(id) ON DELETE CASCADE,
			position INT DEFAULT 0,
			PRIMARY KEY (draft_id, photo_id)
		)`,

		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_photos_user_id ON photos(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_photos_deleted_at ON photos(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_books_user_id ON books(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pages_book_id ON pages(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clusters_user_id ON clusters(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_page_drafts_user_id ON page_drafts(user_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// User operations

// GetOrCreateUser gets or creates a user from Clerk ID
func (db *Database) GetOrCreateUser(ctx context.Context, clerkID, email, name string) (*DBUser, error) {
	var user DBUser

	// Try to get existing user
	err := db.pool.QueryRow(ctx, `
		SELECT id, clerk_id, email, name, created_at, updated_at 
		FROM users WHERE clerk_id = $1
	`, clerkID).Scan(&user.ID, &user.ClerkID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		return &user, nil
	}

	// Create new user
	userID := generateID()
	err = db.pool.QueryRow(ctx, `
		INSERT INTO users (id, clerk_id, email, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, clerk_id, email, name, created_at, updated_at
	`, userID, clerkID, email, name).Scan(&user.ID, &user.ClerkID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Photo operations

// CreatePhoto inserts a new photo record
func (db *Database) CreatePhoto(ctx context.Context, photo *DBPhoto) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO photos (id, user_id, filename, original_filename, gcs_path, thumb_gcs_path, size_bytes, content_type, width, height, taken_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, photo.ID, photo.UserID, photo.Filename, photo.OriginalFilename, photo.GCSPath, photo.ThumbGCSPath, photo.SizeBytes, photo.ContentType, photo.Width, photo.Height, photo.TakenAt)

	return err
}

// GetPhotosByUser returns all photos for a user
func (db *Database) GetPhotosByUser(ctx context.Context, userID string) ([]DBPhoto, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, user_id, filename, original_filename, gcs_path, thumb_gcs_path, size_bytes, content_type, width, height, taken_at, created_at
		FROM photos 
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []DBPhoto
	for rows.Next() {
		var p DBPhoto
		err := rows.Scan(&p.ID, &p.UserID, &p.Filename, &p.OriginalFilename, &p.GCSPath, &p.ThumbGCSPath, &p.SizeBytes, &p.ContentType, &p.Width, &p.Height, &p.TakenAt, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}

	return photos, nil
}

// GetPhotoByID returns a photo by ID
func (db *Database) GetPhotoByID(ctx context.Context, photoID string) (*DBPhoto, error) {
	var p DBPhoto
	err := db.pool.QueryRow(ctx, `
		SELECT id, user_id, filename, original_filename, gcs_path, thumb_gcs_path, size_bytes, content_type, width, height, taken_at, created_at
		FROM photos 
		WHERE id = $1 AND deleted_at IS NULL
	`, photoID).Scan(&p.ID, &p.UserID, &p.Filename, &p.OriginalFilename, &p.GCSPath, &p.ThumbGCSPath, &p.SizeBytes, &p.ContentType, &p.Width, &p.Height, &p.TakenAt, &p.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPhotosByIDs returns photos by their IDs for a specific user
func (db *Database) GetPhotosByIDs(ctx context.Context, userID string, photoIDs []string) ([]DBPhoto, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, user_id, filename, original_filename, gcs_path, thumb_gcs_path, size_bytes, content_type, width, height, taken_at, created_at
		FROM photos 
		WHERE user_id = $1 AND id = ANY($2) AND deleted_at IS NULL
	`, userID, photoIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []DBPhoto
	for rows.Next() {
		var p DBPhoto
		err := rows.Scan(&p.ID, &p.UserID, &p.Filename, &p.OriginalFilename, &p.GCSPath, &p.ThumbGCSPath, &p.SizeBytes, &p.ContentType, &p.Width, &p.Height, &p.TakenAt, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}

	return photos, nil
}

// SoftDeletePhoto marks a photo as deleted
func (db *Database) SoftDeletePhoto(ctx context.Context, userID, photoID string) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE photos SET deleted_at = NOW() 
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, photoID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("photo not found or already deleted")
	}
	return nil
}

// HardDeletePhoto permanently deletes a photo from the database
func (db *Database) HardDeletePhoto(ctx context.Context, photoID string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM photos WHERE id = $1`, photoID)
	return err
}

// HardDeletePhotos permanently deletes multiple photos from the database
func (db *Database) HardDeletePhotos(ctx context.Context, photoIDs []string) error {
	if len(photoIDs) == 0 {
		return nil
	}
	_, err := db.pool.Exec(ctx, `DELETE FROM photos WHERE id = ANY($1)`, photoIDs)
	return err
}

// GetPhotoPathsByIDs returns the GCS paths for photos by their IDs
func (db *Database) GetPhotoPathsByIDs(ctx context.Context, photoIDs []string) ([]struct {
	ID           string
	GCSPath      string
	ThumbGCSPath string
}, error) {
	if len(photoIDs) == 0 {
		return nil, nil
	}
	rows, err := db.pool.Query(ctx, `
		SELECT id, gcs_path, COALESCE(thumb_gcs_path, '') as thumb_gcs_path
		FROM photos WHERE id = ANY($1)
	`, photoIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		ID           string
		GCSPath      string
		ThumbGCSPath string
	}
	for rows.Next() {
		var r struct {
			ID           string
			GCSPath      string
			ThumbGCSPath string
		}
		if err := rows.Scan(&r.ID, &r.GCSPath, &r.ThumbGCSPath); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// GetDraftAllPhotoIDs returns all photo IDs for a draft (from cluster_photos via cluster_id)
func (db *Database) GetDraftAllPhotoIDs(ctx context.Context, draftID string) ([]string, error) {
	// First get the cluster_id for this draft
	var clusterID sql.NullString
	err := db.pool.QueryRow(ctx, `
		SELECT cluster_id FROM page_drafts WHERE id = $1
	`, draftID).Scan(&clusterID)
	if err != nil {
		return nil, err
	}

	if !clusterID.Valid {
		// No cluster associated, just return draft photos
		return db.GetDraftPhotos(ctx, draftID)
	}

	// Get all photos from the cluster (the original set before any discards)
	rows, err := db.pool.Query(ctx, `
		SELECT photo_id FROM cluster_photos WHERE cluster_id = $1
	`, clusterID.String)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photoIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		photoIDs = append(photoIDs, id)
	}
	return photoIDs, nil
}

// Cluster operations

// CreateCluster creates a new photo cluster
func (db *Database) CreateCluster(ctx context.Context, cluster *DBCluster) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO clusters (id, user_id, title, description, theme, date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, cluster.ID, cluster.UserID, cluster.Title, cluster.Description, cluster.Theme, cluster.Date)
	return err
}

// AddPhotosToCluster adds photos to a cluster
func (db *Database) AddPhotosToCluster(ctx context.Context, clusterID string, photoIDs []string) error {
	for _, photoID := range photoIDs {
		_, err := db.pool.Exec(ctx, `
			INSERT INTO cluster_photos (cluster_id, photo_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, clusterID, photoID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Draft operations

// CreateDraft creates a new page draft
func (db *Database) CreateDraft(ctx context.Context, draft *DBPageDraft) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO page_drafts (id, user_id, cluster_id, title, description, theme, background_gcs_path, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, draft.ID, draft.UserID, draft.ClusterID, draft.Title, draft.Description, draft.Theme, draft.BackgroundGCSPath, draft.Status)
	return err
}

// AddPhotosToDraft adds photos to a draft
func (db *Database) AddPhotosToDraft(ctx context.Context, draftID string, photoIDs []string) error {
	for i, photoID := range photoIDs {
		_, err := db.pool.Exec(ctx, `
			INSERT INTO draft_photos (draft_id, photo_id, position)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`, draftID, photoID, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetDraftsByUser returns all drafts for a user
func (db *Database) GetDraftsByUser(ctx context.Context, userID string) ([]DBPageDraft, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, user_id, cluster_id, title, description, theme, background_gcs_path, status, created_at, updated_at
		FROM page_drafts 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []DBPageDraft
	for rows.Next() {
		var d DBPageDraft
		err := rows.Scan(&d.ID, &d.UserID, &d.ClusterID, &d.Title, &d.Description, &d.Theme, &d.BackgroundGCSPath, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, d)
	}

	return drafts, nil
}

// GetDraftByID returns a draft by ID
func (db *Database) GetDraftByID(ctx context.Context, draftID string) (*DBPageDraft, error) {
	var d DBPageDraft
	err := db.pool.QueryRow(ctx, `
		SELECT id, user_id, cluster_id, title, description, theme, background_gcs_path, status, created_at, updated_at
		FROM page_drafts WHERE id = $1
	`, draftID).Scan(&d.ID, &d.UserID, &d.ClusterID, &d.Title, &d.Description, &d.Theme, &d.BackgroundGCSPath, &d.Status, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetDraftPhotos returns photo IDs for a draft
func (db *Database) GetDraftPhotos(ctx context.Context, draftID string) ([]string, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT photo_id FROM draft_photos 
		WHERE draft_id = $1 
		ORDER BY position
	`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photoIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		photoIDs = append(photoIDs, id)
	}
	return photoIDs, nil
}

// UpdateDraft updates a draft
func (db *Database) UpdateDraft(ctx context.Context, draft *DBPageDraft) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE page_drafts 
		SET title = $1, description = $2, theme = $3, background_gcs_path = $4, status = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
	`, draft.Title, draft.Description, draft.Theme, draft.BackgroundGCSPath, draft.Status, draft.ID, draft.UserID)
	return err
}

// DeleteDraft deletes a draft
func (db *Database) DeleteDraft(ctx context.Context, userID, draftID string) error {
	result, err := db.pool.Exec(ctx, `
		DELETE FROM page_drafts WHERE id = $1 AND user_id = $2
	`, draftID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("draft not found")
	}
	return nil
}

// Helper function
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

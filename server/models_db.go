package main

import (
	"database/sql"
	"time"
)

// DBUser represents a user in the database
type DBUser struct {
	ID        string
	ClerkID   string
	Email     sql.NullString
	Name      sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DBPhoto represents a photo in the database
type DBPhoto struct {
	ID               string
	UserID           string
	Filename         string
	OriginalFilename string
	GCSPath          string
	ThumbGCSPath     sql.NullString
	SizeBytes        int64
	ContentType      string
	Width            sql.NullInt32
	Height           sql.NullInt32
	TakenAt          sql.NullTime
	CreatedAt        time.Time
	DeletedAt        sql.NullTime
}

// DBBook represents a memory book in the database
type DBBook struct {
	ID           string
	UserID       string
	Title        string
	Description  sql.NullString
	CoverPhotoID sql.NullString
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DBPage represents a book page in the database
type DBPage struct {
	ID                string
	BookID            string
	PageNumber        int
	Title             sql.NullString
	Description       sql.NullString
	Theme             sql.NullString
	BackgroundGCSPath sql.NullString
	LayoutJSON        sql.NullString
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// DBCluster represents a photo cluster in the database
type DBCluster struct {
	ID          string
	UserID      string
	Title       sql.NullString
	Description sql.NullString
	Theme       sql.NullString
	Date        sql.NullString
	CreatedAt   time.Time
}

// DBPageDraft represents a page draft in the database
type DBPageDraft struct {
	ID                string
	UserID            string
	ClusterID         sql.NullString
	Title             sql.NullString
	Description       sql.NullString
	Theme             sql.NullString
	BackgroundGCSPath sql.NullString
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ToAPIPhoto converts a DBPhoto to the API Photo format with signed URLs
func (p *DBPhoto) ToAPIPhoto(signedURL, thumbSignedURL string) Photo {
	return Photo{
		ID:         p.ID,
		Filename:   p.OriginalFilename,
		Path:       signedURL, // Now a signed URL instead of a path
		Size:       p.SizeBytes,
		UploadedAt: p.CreatedAt,
	}
}

// ToAPIPageDraft converts a DBPageDraft to the API PageDraft format
func (d *DBPageDraft) ToAPIPageDraft(photoIDs []string, backgroundURL string) PageDraft {
	return PageDraft{
		ID:             d.ID,
		ClusterID:      d.ClusterID.String,
		PhotoIds:       photoIDs,
		Title:          d.Title.String,
		Description:    d.Description.String,
		Theme:          d.Theme.String,
		BackgroundPath: backgroundURL, // Now a signed URL
		Status:         d.Status,
		CreatedAt:      d.CreatedAt.Format(time.RFC3339),
	}
}

// ToAPICluster converts a DBCluster to the API PhotoCluster format
func (c *DBCluster) ToAPICluster(photoIDs []string, backgroundURL string) PhotoCluster {
	return PhotoCluster{
		ID:             c.ID,
		PhotoIds:       photoIDs,
		Theme:          c.Theme.String,
		Title:          c.Title.String,
		Description:    c.Description.String,
		Date:           c.Date.String,
		BackgroundPath: backgroundURL,
	}
}

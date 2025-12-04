package main

import "time"

// Photo represents an uploaded photo
type Photo struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
}

// PhotoCluster represents a group of related photos
type PhotoCluster struct {
	ID          string   `json:"id"`
	PhotoIds    []string `json:"photoIds"`
	Theme       string   `json:"theme"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Date        string   `json:"date"`
}

// PageDraft represents a draft page for the memory book
type PageDraft struct {
	ID          string   `json:"id"`
	ClusterID   string   `json:"clusterId"`
	PhotoIds    []string `json:"photoIds"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Theme       string   `json:"theme"`
	Status      string   `json:"status"` // "draft" | "approved" | "rejected"
	CreatedAt   string   `json:"createdAt"`
}

// ClusterRequest is the request body for clustering photos
type ClusterRequest struct {
	PhotoIds []string `json:"photoIds"`
}

// ClusterResponse is the response for clustering photos
type ClusterResponse struct {
	Clusters []PhotoCluster `json:"clusters"`
	Drafts   []PageDraft    `json:"drafts"`
}

// UploadResponse is the response for photo uploads
type UploadResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Photos  []Photo `json:"photos,omitempty"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

package main

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ExtractPhotoDate extracts the date taken from photo EXIF data
// Returns nil if no date is found or if extraction fails
func ExtractPhotoDate(data []byte) *time.Time {
	reader := bytes.NewReader(data)
	x, err := exif.Decode(reader)
	if err != nil {
		log.Printf("[DEBUG] EXIF decode failed: %v", err)
		return nil
	}

	// Try DateTimeOriginal first (when photo was taken)
	dt, err := x.DateTime()
	if err == nil {
		return &dt
	}

	log.Printf("[DEBUG] No EXIF date found: %v", err)
	return nil
}

// ReadFileData reads the entire file into memory and returns a reader for it
// This allows reading the file multiple times (for EXIF extraction and upload)
func ReadFileData(file io.Reader) ([]byte, error) {
	return io.ReadAll(file)
}

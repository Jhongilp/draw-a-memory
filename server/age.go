package main

import (
	"fmt"
	"time"
)

// CalculateAgeString calculates the child's age at a given date
// Returns a human-readable string like "2 weeks", "3 months", "1 year"
func CalculateAgeString(birthday, photoDate time.Time) string {
	if birthday.IsZero() || photoDate.IsZero() {
		return ""
	}

	// If photo is before birthday, return empty
	if photoDate.Before(birthday) {
		return ""
	}

	// Calculate total days
	totalDays := int(photoDate.Sub(birthday).Hours() / 24)

	// First week
	if totalDays < 7 {
		if totalDays == 0 {
			return "Newborn"
		} else if totalDays == 1 {
			return "1 day old"
		}
		return fmt.Sprintf("%d days old", totalDays)
	}

	// Weeks (up to ~2 months)
	if totalDays < 60 {
		weeks := totalDays / 7
		if weeks == 1 {
			return "1 week old"
		}
		return fmt.Sprintf("%d weeks old", weeks)
	}

	// Calculate months and years
	years := 0
	months := 0

	// Simple month/year calculation
	yearDiff := photoDate.Year() - birthday.Year()
	monthDiff := int(photoDate.Month()) - int(birthday.Month())
	dayDiff := photoDate.Day() - birthday.Day()

	if dayDiff < 0 {
		monthDiff--
	}

	if monthDiff < 0 {
		yearDiff--
		monthDiff += 12
	}

	years = yearDiff
	months = monthDiff

	// Format the output
	if years == 0 {
		if months == 1 {
			return "1 month old"
		}
		return fmt.Sprintf("%d months old", months)
	}

	if years == 1 && months == 0 {
		return "1 year old"
	}

	if years == 1 {
		if months == 1 {
			return "1 year, 1 month old"
		}
		return fmt.Sprintf("1 year, %d months old", months)
	}

	if months == 0 {
		return fmt.Sprintf("%d years old", years)
	}

	if months == 1 {
		return fmt.Sprintf("%d years, 1 month old", years)
	}

	return fmt.Sprintf("%d years, %d months old", years, months)
}

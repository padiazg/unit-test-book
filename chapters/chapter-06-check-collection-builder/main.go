package check_collection_builder

import (
	"errors"
	"strings"
)

type Document struct {
	Title   string
	Content string
	Tags    []string
}

type ValidationReport struct {
	Errors   []string
	Warnings []string
	Valid    bool
}

func ValidateDocument(doc *Document) *ValidationReport {
	report := &ValidationReport{Valid: true}

	if doc == nil {
		report.Valid = false
		report.Errors = append(report.Errors, "document is nil")
		return report
	}

	doc.Title = strings.TrimSpace(doc.Title)
	doc.Content = strings.TrimSpace(doc.Content)

	if doc.Title == "" {
		report.Valid = false
		report.Errors = append(report.Errors, "title is required")
	}
	if doc.Content == "" {
		report.Valid = false
		report.Errors = append(report.Errors, "content is required")
	}
	if len(doc.Content) < 10 {
		report.Warnings = append(report.Warnings, "content is too short")
	}

	return report
}

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
)

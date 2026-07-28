// Package pdf renders plain text to PDF bytes. It knows nothing about
// templates, patients or clinics — callers resolve {{tag}} placeholders
// into final text before calling Render, so this stays a pure renderer.
package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// Render lays out title and body on an A4 page — body wraps across
// additional pages automatically via MultiCell — and returns the finished
// PDF bytes.
func Render(title, body string) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.AddPage()

	// Core fonts (Helvetica/Arial) only understand cp1252, not raw UTF-8 —
	// without this translation, Portuguese accents (á, ã, ç, ...) render as
	// garbage or missing glyphs.
	tr := doc.UnicodeTranslatorFromDescriptor("")

	doc.SetFont("Helvetica", "B", 14)
	doc.MultiCell(0, 8, tr(title), "", "C", false)
	doc.Ln(6)

	doc.SetFont("Helvetica", "", 11)
	doc.MultiCell(0, 6, tr(body), "", "L", false)

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: render: %w", err)
	}
	return buf.Bytes(), nil
}

// Package pdf renders plain text (plus optional letterhead blocks) to PDF
// bytes. It knows nothing about templates, patients or clinics — callers
// resolve {{tag}} placeholders into final text and assemble a RenderOptions
// from tenant/professional data before calling Render, so this stays a pure
// renderer.
package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

const (
	pageMarginMM   = 15.0
	headerTopMM    = 10.0 // where the header block starts drawing from the top
	footerHeightMM = 20.0 // reserved bottom margin when a footer is drawn
)

// RenderOptions controls the optional letterhead blocks drawn around the
// template body — independent of {{tag}} substitution, which the caller has
// already resolved into body before calling Render. Every field is optional;
// an empty/nil value just means that piece of a block is skipped, not that
// rendering fails.
type RenderOptions struct {
	IncludeHeader    bool
	IncludeFooter    bool
	IncludeSignature bool
	// IncludeStamp only has a visible effect when IncludeSignature is also
	// true — the stamp (professional name + registration) is drawn below
	// the signature line, not as an independent block.
	IncludeStamp bool

	ClinicName     string
	ClinicDocument string // CNPJ/CPF, printed as-is
	ClinicAddress  string // one pre-joined line; caller drops empty parts
	ClinicPhone    string
	ClinicEmail    string
	LogoBytes      []byte // nil when no logo is configured/available
	LogoMimeType   string // e.g. "image/png" — tells fpdf the image format

	SignatureCity string // usually the clinic's city; blank prints just the date
	SignatureDate string // pre-formatted, e.g. "4 de agosto de 2026"

	ProfessionalName         string // stamp text is skipped entirely if this is empty
	ProfessionalRegistration string // e.g. "CRM/SP 123456" — line omitted if empty
}

// Render lays out title and body on an A4 page — body wraps across
// additional pages automatically via MultiCell — plus whichever letterhead
// blocks opts enables, and returns the finished PDF bytes.
func Render(title, body string, opts RenderOptions) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")

	// Core fonts (Helvetica/Arial) only understand cp1252, not raw UTF-8 —
	// without this translation, Portuguese accents (á, ã, ç, ...) render as
	// garbage or missing glyphs.
	tr := doc.UnicodeTranslatorFromDescriptor("")

	topMargin := pageMarginMM
	if opts.IncludeHeader {
		// beginpage() sets the cursor to tMargin *before* calling
		// headerFnc, so this must stay small (near the physical top) — the
		// header function itself pushes the cursor down as it draws, and
		// wherever it ends up becomes the effective start of the body.
		topMargin = headerTopMM
	}
	bottomMargin := pageMarginMM
	if opts.IncludeFooter {
		bottomMargin = footerHeightMM
	}
	doc.SetMargins(pageMarginMM, topMargin, pageMarginMM)
	doc.SetAutoPageBreak(true, bottomMargin)

	if opts.IncludeHeader {
		doc.SetHeaderFunc(func() { drawLetterheadHeader(doc, tr, opts) })
	}
	if opts.IncludeFooter {
		doc.SetFooterFunc(func() { drawLetterheadFooter(doc, tr, opts) })
	}

	doc.AddPage()

	doc.SetFont("Helvetica", "B", 14)
	doc.MultiCell(0, 8, tr(title), "", "C", false)
	doc.Ln(6)

	doc.SetFont("Helvetica", "", 11)
	doc.MultiCell(0, 6, tr(body), "", "L", false)

	if opts.IncludeSignature {
		drawSignatureBlock(doc, tr, opts)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: render: %w", err)
	}
	return buf.Bytes(), nil
}

// drawLetterheadHeader draws a centered logo (if present), clinic name, and
// CNPJ/address line, followed by a thin rule — then leaves the cursor
// positioned for the body to start right below it (see the topMargin
// comment in Render for why this relies on ending in the right place
// rather than an explicit reset).
func drawLetterheadHeader(doc *fpdf.Fpdf, tr func(string) string, opts RenderOptions) {
	y := headerTopMM
	pageW, _ := doc.GetPageSize()

	if len(opts.LogoBytes) > 0 {
		if logoHeightUsed := drawLogoSafely(doc, opts, pageW, y); logoHeightUsed > 0 {
			y += logoHeightUsed + 2
		}
	}

	if opts.ClinicName != "" {
		doc.SetY(y)
		doc.SetFont("Helvetica", "B", 12)
		doc.CellFormat(0, 5, tr(opts.ClinicName), "", 2, "C", false, 0, "")
		y = doc.GetY()
	}

	if details := joinNonEmpty(" — ", opts.ClinicDocument, opts.ClinicAddress); details != "" {
		doc.SetFont("Helvetica", "", 8)
		doc.CellFormat(0, 4, tr(details), "", 2, "C", false, 0, "")
		y = doc.GetY()
	}

	doc.SetDrawColor(180, 180, 180)
	doc.Line(pageMarginMM, y+1, pageW-pageMarginMM, y+1)
	doc.SetY(y + 5)
}

// drawLogoSafely registers and places the logo image, returning the height
// it occupied (0 if nothing was drawn). A malformed/unsupported logo must
// never take down the rest of the document, but fpdf can fail that in two
// different ways: some inputs make its image decoders panic outright
// (recover() below), while others just set fpdf's own internal error flag
// (f.err) — which, left alone, silently no-ops every fpdf call after it,
// including Output(), so the *whole* PDF would fail even though only the
// logo was bad. ClearError() resets that flag so title/body/footer/
// signature still render normally. Both failure modes were hit live before
// this guard existed.
func drawLogoSafely(doc *fpdf.Fpdf, opts RenderOptions, pageW, y float64) (heightUsed float64) {
	defer func() {
		if r := recover(); r != nil {
			heightUsed = 0
		}
	}()

	imgType := doc.ImageTypeFromMime(opts.LogoMimeType)
	if imgType == "" {
		imgType = "PNG"
	}
	info := doc.RegisterImageOptionsReader(
		"clinic-logo",
		fpdf.ImageOptions{ImageType: imgType},
		bytes.NewReader(opts.LogoBytes),
	)
	if doc.Err() {
		doc.ClearError()
		return 0
	}
	if info == nil {
		return 0
	}
	const logoHeight = 14.0
	logoWidth := logoHeight
	if h := info.Height(); h > 0 {
		logoWidth = info.Width() * (logoHeight / h)
	}
	doc.ImageOptions("clinic-logo", (pageW-logoWidth)/2, y, logoWidth, logoHeight, false,
		fpdf.ImageOptions{ImageType: imgType}, 0, "")
	if doc.Err() {
		doc.ClearError()
		return 0
	}
	return logoHeight
}

// drawLetterheadFooter is invoked by fpdf on every page (see
// SetFooterFunc) — SetY with a negative value is fpdf's own convention for
// "measured from the bottom of the page", so this always lands in the
// reserved bottomMargin regardless of page count.
func drawLetterheadFooter(doc *fpdf.Fpdf, tr func(string) string, opts RenderOptions) {
	contact := joinNonEmpty("  |  ", opts.ClinicPhone, opts.ClinicEmail)
	if contact == "" {
		return
	}
	pageW, _ := doc.GetPageSize()
	doc.SetY(-footerHeightMM + 6)
	doc.SetDrawColor(180, 180, 180)
	doc.Line(pageMarginMM, doc.GetY(), pageW-pageMarginMM, doc.GetY())
	doc.Ln(3)
	doc.SetFont("Helvetica", "", 8)
	doc.CellFormat(0, 4, tr(contact), "", 0, "C", false, 0, "")
}

// drawSignatureBlock prints "[Cidade], [Data]" above a dashed line, with
// the professional's name/registration below it when IncludeStamp is set —
// always at the end of the body, once, never repeated per page (unlike
// header/footer).
func drawSignatureBlock(doc *fpdf.Fpdf, tr func(string) string, opts RenderOptions) {
	doc.Ln(14)
	if dateLine := joinNonEmpty(", ", opts.SignatureCity, opts.SignatureDate); dateLine != "" {
		doc.SetFont("Helvetica", "", 11)
		doc.CellFormat(0, 6, tr(dateLine), "", 2, "C", false, 0, "")
	}

	doc.Ln(16)
	pageW, _ := doc.GetPageSize()
	const lineWidth = 70.0
	x := (pageW - lineWidth) / 2
	y := doc.GetY()
	doc.SetDrawColor(0, 0, 0)
	doc.SetDashPattern([]float64{1, 1}, 0)
	doc.Line(x, y, x+lineWidth, y)
	doc.SetDashPattern(nil, 0) // restore solid lines for anything drawn after
	doc.Ln(4)

	if opts.IncludeStamp && opts.ProfessionalName != "" {
		doc.SetFont("Helvetica", "B", 10)
		doc.CellFormat(0, 5, tr(opts.ProfessionalName), "", 2, "C", false, 0, "")
		if opts.ProfessionalRegistration != "" {
			doc.SetFont("Helvetica", "", 9)
			doc.CellFormat(0, 5, tr(opts.ProfessionalRegistration), "", 2, "C", false, 0, "")
		}
	}
}

func joinNonEmpty(sep string, parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, sep)
}

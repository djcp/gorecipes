package export

import (
	"bytes"
	"strings"

	"github.com/djcp/gorecipes/internal/models"
	"github.com/djcp/gorecipes/internal/version"
	"github.com/go-pdf/fpdf"
)

// ToPDF renders a recipe as a PDF document (Letter size).
func ToPDF(r *models.Recipe, opts Options) ([]byte, error) {
	f := fpdf.New("P", "mm", "Letter", "")
	// tr converts UTF-8 strings to cp1252 (the encoding used by core PDF fonts
	// like Helvetica). Without this, multi-byte UTF-8 sequences for characters
	// such as • (U+2022) and ° (U+00B0) are misread as two Latin-1 characters,
	// producing mojibake like "â€¢" and "Â°".
	tr := f.UnicodeTranslatorFromDescriptor("")
	f.SetMargins(20, 25, 20)
	f.AddPage()

	totalW, _ := f.GetPageSize()
	pw := totalW - 40 // printable width (subtract left + right margins)

	// Title
	f.SetFont("Helvetica", "B", 18)
	f.SetTextColor(201, 100, 66) // terracotta
	f.MultiCell(pw, 10, tr(r.Name), "", "L", false)
	f.Ln(2)

	// Timing / servings subtitle
	f.SetFont("Helvetica", "", 10)
	f.SetTextColor(142, 129, 120) // warm gray
	var meta []string
	if t := r.TimingSummary(); t != "" {
		meta = append(meta, t)
	}
	if r.Servings != nil && *r.Servings > 0 {
		units := r.ServingUnits
		if units == "" {
			units = "servings"
		}
		meta = append(meta, formatServings(*r.Servings, units))
	}
	if len(meta) > 0 {
		f.MultiCell(pw, 6, tr(strings.Join(meta, "  \u00b7  ")), "", "L", false)
	}

	// Tags
	f.SetFont("Helvetica", "", 9)
	for _, ctx := range models.AllTagContexts {
		tags := r.TagsByContext(ctx)
		if len(tags) > 0 {
			f.MultiCell(pw, 5, tr(TagContextLabel(ctx)+": "+strings.Join(tags, ", ")), "", "L", false)
		}
	}
	f.Ln(4)

	// Description
	if r.Description != "" {
		f.SetFont("Helvetica", "I", 11)
		f.SetTextColor(92, 74, 60) // dark warm brown
		f.MultiCell(pw, 6, tr(r.Description), "", "L", false)
		f.Ln(4)
	}

	// Ingredients section
	if len(r.Ingredients) > 0 {
		renderPDFSection(f, tr, "Ingredients", pw)
		f.SetFont("Helvetica", "", 11)
		f.SetTextColor(50, 50, 50)
		currentSection := ""
		for _, ing := range r.Ingredients {
			if ing.Section != currentSection && ing.Section != "" {
				f.Ln(2)
				f.SetFont("Helvetica", "B", 11)
				f.MultiCell(pw, 6, tr(ing.Section), "", "L", false)
				f.SetFont("Helvetica", "", 11)
				currentSection = ing.Section
			}
			f.MultiCell(pw, 6, tr("  \u2022  "+ing.DisplayString()), "", "L", false)
		}
		f.Ln(4)
	}

	// Directions section
	if r.Directions != "" {
		renderPDFSection(f, tr, "Directions", pw)
		f.SetFont("Helvetica", "", 11)
		f.SetTextColor(50, 50, 50)
		f.MultiCell(pw, 6, tr(r.Directions), "", "L", false)
	}

	// Source URL
	if r.SourceURL != "" {
		f.Ln(6)
		f.SetFont("Helvetica", "", 9)
		f.SetTextColor(142, 129, 120)
		f.MultiCell(pw, 5, tr("Source: "+r.SourceURL), "", "L", false)
	}

	// Footer: credits left, version right, separated by a thin rule.
	f.Ln(8)
	f.SetDrawColor(200, 200, 200)
	f.SetLineWidth(0.2)
	f.Line(f.GetX(), f.GetY(), f.GetX()+pw, f.GetY())
	f.Ln(3)
	f.SetFont("Helvetica", "I", 8)
	f.SetTextColor(128, 128, 128)
	versionStr := "exported from gorecipes " + version.Version
	if opts.Credits != "" {
		halfW := pw / 2
		f.CellFormat(halfW, 5, tr(opts.Credits), "", 0, "L", false, 0, "")
		f.CellFormat(halfW, 5, tr(versionStr), "", 1, "R", false, 0, "")
	} else {
		f.MultiCell(pw, 5, tr(versionStr), "", "R", false)
	}

	var buf bytes.Buffer
	if err := f.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderPDFSection(f *fpdf.Fpdf, tr func(string) string, title string, pw float64) {
	f.SetFont("Helvetica", "B", 13)
	f.SetTextColor(124, 158, 110) // sage green
	f.MultiCell(pw, 8, tr(title), "", "L", false)
	// Thin horizontal rule
	f.SetDrawColor(220, 213, 204) // light warm gray
	f.SetLineWidth(0.3)
	x := f.GetX()
	y := f.GetY()
	f.Line(x, y, x+pw, y)
	f.Ln(3)
}

package export

import (
	"strings"

	"github.com/djcp/gorecipes/internal/models"
)

// ToText renders a recipe as plain text.
func ToText(r *models.Recipe) string {
	var sb strings.Builder

	// Title + underline
	sb.WriteString(r.Name + "\n")
	sb.WriteString(strings.Repeat("=", len([]rune(r.Name))) + "\n")

	// Timing / servings
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
		sb.WriteString(strings.Join(meta, "  ·  ") + "\n")
	}

	// Tags
	for _, ctx := range models.AllTagContexts {
		tags := r.TagsByContext(ctx)
		if len(tags) > 0 {
			sb.WriteString(TagContextLabel(ctx) + ": " + strings.Join(tags, ", ") + "\n")
		}
	}

	// Description
	if r.Description != "" {
		sb.WriteString("\n" + r.Description + "\n")
	}

	// Ingredients
	if len(r.Ingredients) > 0 {
		sb.WriteString("\nINGREDIENTS\n")
		sb.WriteString("-----------\n")
		currentSection := ""
		for _, ing := range r.Ingredients {
			if ing.Section != currentSection && ing.Section != "" {
				sb.WriteString("\n  " + ing.Section + "\n")
				currentSection = ing.Section
			}
			sb.WriteString("  " + ing.DisplayString() + "\n")
		}
	}

	// Directions
	if r.Directions != "" {
		sb.WriteString("\nDIRECTIONS\n")
		sb.WriteString("----------\n")
		sb.WriteString(r.Directions + "\n")
	}

	// Source
	if r.SourceURL != "" {
		sb.WriteString("\nSource: " + r.SourceURL + "\n")
	}

	return sb.String()
}

func formatServings(n int, units string) string {
	if n == 1 {
		return "Makes 1 " + units
	}
	return "Makes " + itoa(n) + " " + units
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

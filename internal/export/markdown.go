package export

import (
	"fmt"
	"strings"

	"github.com/djcp/gorecipes/internal/models"
	"github.com/djcp/gorecipes/internal/version"
)

// ToMarkdown renders a recipe as a Markdown document.
func ToMarkdown(r *models.Recipe) string {
	var sb strings.Builder

	sb.WriteString("# " + r.Name + "\n\n")

	// Timing / servings line
	var meta []string
	if r.PreparationTime != nil && *r.PreparationTime > 0 {
		meta = append(meta, fmt.Sprintf("**Prep:** %s", FormatMins(*r.PreparationTime)))
	}
	if r.CookingTime != nil && *r.CookingTime > 0 {
		meta = append(meta, fmt.Sprintf("**Cook:** %s", FormatMins(*r.CookingTime)))
	}
	if r.Servings != nil && *r.Servings > 0 {
		units := r.ServingUnits
		if units == "" {
			units = "servings"
		}
		meta = append(meta, fmt.Sprintf("**Serves:** %d %s", *r.Servings, units))
	}
	if len(meta) > 0 {
		sb.WriteString(strings.Join(meta, " | ") + "\n\n")
	}

	// Description
	if r.Description != "" {
		sb.WriteString("> " + r.Description + "\n\n")
	}

	// Tags as blockquote lines
	for _, ctx := range models.AllTagContexts {
		tags := r.TagsByContext(ctx)
		if len(tags) > 0 {
			sb.WriteString("> **" + TagContextLabel(ctx) + ":** " + strings.Join(tags, ", ") + "\n")
		}
	}
	sb.WriteString("\n")

	// Ingredients
	if len(r.Ingredients) > 0 {
		sb.WriteString("## Ingredients\n\n")
		currentSection := ""
		for _, ing := range r.Ingredients {
			if ing.Section != currentSection && ing.Section != "" {
				sb.WriteString("\n### " + ing.Section + "\n\n")
				currentSection = ing.Section
			}
			sb.WriteString("- " + ing.DisplayString() + "\n")
		}
		sb.WriteString("\n")
	}

	// Directions
	if r.Directions != "" {
		sb.WriteString("## Directions\n\n")
		sb.WriteString(r.Directions + "\n")
	}

	// Source
	if r.SourceURL != "" {
		sb.WriteString("\n---\n\nSource: " + r.SourceURL + "\n")
	}

	// Attribution footer — right-aligned, small text
	sb.WriteString("\n<p align=\"right\"><sub>exported from gorecipes " + version.Version + "</sub></p>\n")

	return sb.String()
}

package export

import (
	"fmt"
	"strings"

	"github.com/djcp/gorecipes/internal/models"
)

// ToRTF renders a recipe as an RTF 1.x document.
func ToRTF(r *models.Recipe) string {
	var sb strings.Builder

	sb.WriteString("{\\rtf1\\ansi\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}}\n")
	sb.WriteString("{\\colortbl;\\red201\\green100\\blue66;\\red124\\green158\\blue110;\\red142\\green129\\blue120;}\n")
	sb.WriteString("\\f0\\fs22\n")

	// Title
	sb.WriteString(fmt.Sprintf("{\\fs36\\b\\cf1 %s\\cf0\\b0\\par}\n", rtfEscape(r.Name)))
	sb.WriteString("\\par\n")

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
		sb.WriteString(fmt.Sprintf("{\\fs20\\cf3 %s\\cf0\\par}\n", rtfEscape(strings.Join(meta, "  \u00b7  "))))
	}

	// Tags
	for _, ctx := range models.AllTagContexts {
		tags := r.TagsByContext(ctx)
		if len(tags) > 0 {
			label := TagContextLabel(ctx)
			sb.WriteString(fmt.Sprintf("{\\fs18\\cf3 %s: %s\\cf0\\par}\n",
				rtfEscape(label), rtfEscape(strings.Join(tags, ", "))))
		}
	}
	sb.WriteString("\\par\n")

	// Description
	if r.Description != "" {
		sb.WriteString(fmt.Sprintf("{\\fs22\\i %s\\i0\\par}\n", rtfEscape(r.Description)))
		sb.WriteString("\\par\n")
	}

	// Ingredients
	if len(r.Ingredients) > 0 {
		sb.WriteString("{\\fs26\\b\\cf2 Ingredients\\cf0\\b0\\par}\n")
		sb.WriteString("\\par\n")
		currentSection := ""
		for _, ing := range r.Ingredients {
			if ing.Section != currentSection && ing.Section != "" {
				sb.WriteString(fmt.Sprintf("{\\fs22\\b %s\\b0\\par}\n", rtfEscape(ing.Section)))
				currentSection = ing.Section
			}
			sb.WriteString(fmt.Sprintf("{\\fs22 - %s\\par}\n", rtfEscape(ing.DisplayString())))
		}
		sb.WriteString("\\par\n")
	}

	// Directions
	if r.Directions != "" {
		sb.WriteString("{\\fs26\\b\\cf2 Directions\\cf0\\b0\\par}\n")
		sb.WriteString("\\par\n")
		sb.WriteString(fmt.Sprintf("{\\fs22 %s\\par}\n", rtfEscape(r.Directions)))
		sb.WriteString("\\par\n")
	}

	// Source URL
	if r.SourceURL != "" {
		sb.WriteString(fmt.Sprintf("{\\fs18\\cf3 Source: %s\\cf0\\par}\n", rtfEscape(r.SourceURL)))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// rtfEscape escapes the three RTF control characters in a string.
func rtfEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "{", `\{`)
	s = strings.ReplaceAll(s, "}", `\}`)
	return s
}

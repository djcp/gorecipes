package export

import (
	"fmt"
	"strings"

	"github.com/djcp/gorecipes/internal/models"
	"github.com/djcp/gorecipes/internal/version"
)

// ToRTF renders a recipe as an RTF 1.x document (cp1252 encoding).
func ToRTF(r *models.Recipe) string {
	var sb strings.Builder

	// \ansicpg1252 declares the default code page explicitly so RTF readers
	// use cp1252 rather than the system default when interpreting \'XX escapes.
	sb.WriteString("{\\rtf1\\ansi\\ansicpg1252\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}}\n")
	// cf1=terracotta  cf2=sage green  cf3=warm gray  cf4=50% gray (attribution)
	sb.WriteString("{\\colortbl;\\red201\\green100\\blue66;\\red124\\green158\\blue110;\\red142\\green129\\blue120;\\red128\\green128\\blue128;}\n")
	sb.WriteString("\\f0\\fs22\n")

	// Title
	sb.WriteString(fmt.Sprintf("{\\fs36\\b\\cf1 %s\\cf0\\b0\\par}\n", rtfEnc(r.Name)))
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
		sb.WriteString(fmt.Sprintf("{\\fs20\\cf3 %s\\cf0\\par}\n", rtfEnc(strings.Join(meta, "  \u00b7  "))))
	}

	// Tags
	for _, ctx := range models.AllTagContexts {
		tags := r.TagsByContext(ctx)
		if len(tags) > 0 {
			label := TagContextLabel(ctx)
			sb.WriteString(fmt.Sprintf("{\\fs18\\cf3 %s: %s\\cf0\\par}\n",
				rtfEnc(label), rtfEnc(strings.Join(tags, ", "))))
		}
	}
	sb.WriteString("\\par\n")

	// Description
	if r.Description != "" {
		sb.WriteString(fmt.Sprintf("{\\fs22\\i %s\\i0\\par}\n", rtfEnc(r.Description)))
		sb.WriteString("\\par\n")
	}

	// Ingredients
	if len(r.Ingredients) > 0 {
		sb.WriteString("{\\fs26\\b\\cf2 Ingredients\\cf0\\b0\\par}\n")
		sb.WriteString("\\par\n")
		currentSection := ""
		for _, ing := range r.Ingredients {
			if ing.Section != currentSection && ing.Section != "" {
				sb.WriteString(fmt.Sprintf("{\\fs22\\b %s\\b0\\par}\n", rtfEnc(ing.Section)))
				currentSection = ing.Section
			}
			sb.WriteString(fmt.Sprintf("{\\fs22 - %s\\par}\n", rtfEnc(ing.DisplayString())))
		}
		sb.WriteString("\\par\n")
	}

	// Directions
	if r.Directions != "" {
		sb.WriteString("{\\fs26\\b\\cf2 Directions\\cf0\\b0\\par}\n")
		sb.WriteString("\\par\n")
		sb.WriteString(fmt.Sprintf("{\\fs22 %s\\par}\n", rtfEnc(r.Directions)))
		sb.WriteString("\\par\n")
	}

	// Source URL
	if r.SourceURL != "" {
		sb.WriteString(fmt.Sprintf("{\\fs18\\cf3 Source: %s\\cf0\\par}\n", rtfEnc(r.SourceURL)))
	}

	// Attribution footer — right-aligned, 50% gray
	sb.WriteString(fmt.Sprintf("{\\pard\\qr\\fs16\\cf4 exported from gorecipes %s\\cf0\\par}\n",
		rtfEnc(version.Version)))

	sb.WriteString("}\n")
	return sb.String()
}

// cp1252Special maps Unicode code points in the U+0080–U+009F range that
// cp1252 assigns to printable characters (unlike Latin-1, which uses them for
// control codes) to their cp1252 byte values.
var cp1252Special = map[rune]byte{
	0x20AC: 0x80, // €
	0x201A: 0x82, // ‚
	0x0192: 0x83, // ƒ
	0x201E: 0x84, // „
	0x2026: 0x85, // …
	0x2020: 0x86, // †
	0x2021: 0x87, // ‡
	0x02C6: 0x88, // ˆ
	0x2030: 0x89, // ‰
	0x0160: 0x8A, // Š
	0x2039: 0x8B, // ‹
	0x0152: 0x8C, // Œ
	0x017D: 0x8E, // Ž
	0x2018: 0x91, // '
	0x2019: 0x92, // '
	0x201C: 0x93, // "
	0x201D: 0x94, // "
	0x2022: 0x95, // •
	0x2013: 0x96, // –
	0x2014: 0x97, // —
	0x02DC: 0x98, // ˜
	0x2122: 0x99, // ™
	0x0161: 0x9A, // š
	0x203A: 0x9B, // ›
	0x0153: 0x9C, // œ
	0x017E: 0x9E, // ž
	0x0178: 0x9F, // Ÿ
}

// rtfEnc encodes a UTF-8 string for safe embedding in an RTF 1.x document
// declared with \ansi\ansicpg1252. It handles characters as follows:
//
//   - RTF metacharacters (\, {, }): escaped as \\, \{, \}
//   - Newlines (\n): converted to \par (paragraph break)
//   - \r: dropped (line endings are handled by \n)
//   - Other control characters (< 0x20): dropped
//   - ASCII printable (0x20–0x7E): passed through unchanged
//   - Latin-1 supplement (U+00A0–U+00FF): \'XX, same byte value as cp1252
//   - cp1252 special range (e.g. •, –, °, curly quotes): \'XX via cp1252Special
//   - Everything else: \uN? (signed 16-bit decimal Unicode escape, fallback '?')
func rtfEnc(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r == '\\':
			sb.WriteString(`\\`)
		case r == '{':
			sb.WriteString(`\{`)
		case r == '}':
			sb.WriteString(`\}`)
		case r == '\n':
			sb.WriteString(`\par` + "\n")
		case r == '\r':
			// dropped; \n handles line endings
		case r < 0x20:
			// drop other control characters
		case r < 0x80:
			sb.WriteRune(r)
		case r >= 0xA0 && r <= 0xFF:
			// Latin-1 supplement: code point equals cp1252 byte value
			fmt.Fprintf(&sb, "\\'%02x", byte(r))
		default:
			if b, ok := cp1252Special[r]; ok {
				fmt.Fprintf(&sb, "\\'%02x", b)
			} else {
				// RTF Unicode escape: signed 16-bit decimal with '?' fallback
				n := int32(r)
				if n > 32767 {
					n -= 65536
				}
				fmt.Fprintf(&sb, `\u%d?`, n)
			}
		}
	}
	return sb.String()
}

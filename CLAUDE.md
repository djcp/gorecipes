# gorecipes — development notes

## UI / lipgloss rendering

### Centering multi-line blocks (dialogs, forms, overlays)

Never use `strings.Repeat(" ", leftPad) + block` to center a multi-line lipgloss-rendered string.
That only pads the **first** line; every subsequent line starts at column 0.

Always use `lipgloss.PlaceHorizontal`:

```go
sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, block))
```

This applies to any multi-line element: confirmation dialogs, bordered form inputs, info boxes, overlays — anything that spans more than one terminal line and needs to be centered.

### Left-indenting multi-line blocks (form inputs, bordered bars)

Never write a manual indent string before a multi-line lipgloss block:

```go
// WRONG — only the first line gets the indent
sb.WriteString("  ")
sb.WriteString(bar)
```

Use `MarginLeft` in the lipgloss style instead so every line is indented consistently:

```go
bar := lipgloss.NewStyle().
    Border(...).
    Width(m.width - 6).
    MarginLeft(2).   // ← lipgloss applies this to all lines
    Render(content)
sb.WriteString(bar)
```

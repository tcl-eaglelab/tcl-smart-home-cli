package output

import (
	"fmt"
	"os"
	"strings"
)

var noColor = os.Getenv("NO_COLOR") != ""

func color(code string) string {
	if noColor {
		return ""
	}
	return code
}

var (
	colorReset  = color("\033[0m")
	colorRed    = color("\033[31m")
	colorGreen  = color("\033[32m")
	colorYellow = color("\033[33m")
	colorBlue   = color("\033[34m")
	colorCyan   = color("\033[36m")
	colorGray   = color("\033[90m")
	colorBold   = color("\033[1m")
)

// Success prints a green success message to stdout.
func Success(msg string) {
	fmt.Println(colorGreen + "✔" + colorReset + " " + msg)
}

// Error prints a red error message to stderr.
func Error(msg string) {
	fmt.Fprintln(os.Stderr, colorRed+"✖"+colorReset+" "+msg)
}

// Warn prints a warning message to stderr.
func Warn(msg string) {
	fmt.Fprintln(os.Stderr, colorYellow+"⚠"+colorReset+" "+msg)
}

// Info prints a blue informational message to stdout.
func Info(msg string) {
	fmt.Println(colorBlue + "ℹ" + colorReset + " " + msg)
}

// Gray prints a gray (dim) message to stdout.
func Gray(msg string) {
	fmt.Println(colorGray + msg + colorReset)
}

// Bold prints a bold message to stdout.
func Bold(msg string) {
	fmt.Println(colorBold + msg + colorReset)
}

// FormatCyanBold returns s wrapped in bold cyan ANSI codes.
func FormatCyanBold(s string) string { return colorCyan + colorBold + s + colorReset }

// FormatCyan returns s wrapped in cyan ANSI codes.
func FormatCyan(s string) string { return colorCyan + s + colorReset }

// FormatGray returns s wrapped in gray ANSI codes.
func FormatGray(s string) string { return colorGray + s + colorReset }

// StatusBadge returns a colored online/offline badge.
func StatusBadge(online bool) string {
	if online {
		return colorGreen + "● online" + colorReset
	}
	return colorGray + "○ offline" + colorReset
}

// Table prints a formatted table with headers and rows.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runeWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				w := runeWidth(stripAnsi(cell))
				if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	// Print
	printBorder(widths, "┌", "┬", "┐")
	printRow(headers, widths, true)
	printBorder(widths, "├", "┼", "┤")
	for i, row := range rows {
		printRow(row, widths, false)
		if i < len(rows)-1 {
			printBorder(widths, "├", "┼", "┤")
		}
	}
	printBorder(widths, "└", "┴", "┘")
}

func printBorder(widths []int, left, mid, right string) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w+2)
	}
	fmt.Println(left + strings.Join(parts, mid) + right)
}

func printRow(cells []string, widths []int, isHeader bool) {
	parts := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		stripped := stripAnsi(cell)
		pad := widths[i] - runeWidth(stripped)
		if pad < 0 {
			pad = 0
		}
		content := cell + strings.Repeat(" ", pad)
		if isHeader {
			content = colorCyan + colorBold + stripped + colorReset + strings.Repeat(" ", pad)
		}
		parts[i] = " " + content + " "
	}
	fmt.Println("│" + strings.Join(parts, "│") + "│")
}

func runeWidth(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x1100 && isWide(r) {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func isWide(r rune) bool {
	return (r >= 0x1100 && r <= 0x115F) ||
		(r >= 0x2E80 && r <= 0x303E) ||
		(r >= 0x3040 && r <= 0x33BF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x4E00 && r <= 0xA4CF) ||
		(r >= 0xAC00 && r <= 0xD7AF) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF01 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x2FFFD)
}

func stripAnsi(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 'A' || s[j] > 'Z') && (s[j] < 'a' || s[j] > 'z') {
				j++
			}
			if j < len(s) {
				i = j + 1
				continue
			}
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}

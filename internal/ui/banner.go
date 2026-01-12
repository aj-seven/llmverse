package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const bannerText = `
 ██╗      ██╗      ███╗   ███╗ ██╗   ██╗
 ██║      ██║      ████╗ ████║ ██║   ██║
 ██║      ██║      ██╔████╔██║ ██║   ██║
 ██║      ██║      ██║╚██╔╝██║ ╚██╗ ██╔╝
 ███████╗ ███████╗ ██║ ╚═╝ ██║  ╚████╔╝
 ╚══════╝ ╚══════╝ ╚═╝     ╚═╝   ╚═══╝
`

// RenderBanner renders a centered gradient banner.
func RenderBanner(width, height int) string {
	gradient := []string{
		"#00C6FF",
		"#0072FF",
	}

	styledBanner := applyGradient(bannerText, gradient)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		styledBanner,
	)
}

// applyGradient applies a smooth horizontal gradient across each line.
func applyGradient(text string, colors []string) string {
	lines := strings.Split(text, "\n")
	var out []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}

		runes := []rune(line)
		var styled strings.Builder

		for i, r := range runes {
			t := float64(i) / float64(len(runes)-1)
			color := interpolateGradient(colors, t)

			styled.WriteString(
				lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
					Bold(true).
					Render(string(r)),
			)
		}

		out = append(out, styled.String())
	}

	return strings.Join(out, "\n")
}

// interpolateGradient interpolates across multiple color stops.
func interpolateGradient(colors []string, t float64) string {
	if t <= 0 {
		return colors[0]
	}
	if t >= 1 {
		return colors[len(colors)-1]
	}

	segment := t * float64(len(colors)-1)
	i := int(segment)
	localT := segment - float64(i)

	r1, g1, b1 := hexToRGB(colors[i])
	r2, g2, b2 := hexToRGB(colors[i+1])

	return rgbToHex(
		lerp(r1, r2, localT),
		lerp(g1, g2, localT),
		lerp(b1, b2, localT),
	)
}

// Helpers

func lerp(a, b int, t float64) int {
	return int(float64(a) + (float64(b)-float64(a))*t)
}

func hexToRGB(hex string) (int, int, int) {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

package dashboard

import (
	"fmt"
	"io"
	"strings"

	"github.com/unbound-force/unbound-force/internal/metrics"
)

// Sparkline characters for rendering trends.
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// BarChartPoint represents a single bar in a bar chart.
type BarChartPoint struct {
	Label string
	Value float64
}

// RenderBarChart renders an ASCII bar chart to the writer.
func RenderBarChart(title string, data []BarChartPoint, w io.Writer) error {
	if len(data) == 0 {
		return nil
	}

	_, _ = fmt.Fprintf(w, "%s\n", title)
	_, _ = fmt.Fprintf(w, "%s\n", strings.Repeat("─", len(title)))

	maxVal := 0.0
	maxLabel := 0
	for _, d := range data {
		if d.Value > maxVal {
			maxVal = d.Value
		}
		if len(d.Label) > maxLabel {
			maxLabel = len(d.Label)
		}
	}

	barWidth := 30
	for _, d := range data {
		width := 0
		if maxVal > 0 {
			width = int(d.Value / maxVal * float64(barWidth))
		}
		if width < 1 && d.Value > 0 {
			width = 1
		}
		bar := strings.Repeat("█", width)
		_, _ = fmt.Fprintf(w, "  %-*s  %s  %.0f\n", maxLabel, d.Label, bar, d.Value)
	}

	return nil
}

// RenderSparkline renders a Unicode sparkline to the writer.
func RenderSparkline(title string, data []float64, w io.Writer) error {
	if len(data) == 0 {
		return nil
	}

	_, _ = fmt.Fprintf(w, "%s\n", title)
	_, _ = fmt.Fprintf(w, "%s\n", strings.Repeat("─", len(title)))

	minVal := data[0]
	maxVal := data[0]
	sum := 0.0
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
		sum += v
	}
	avg := sum / float64(len(data))

	var spark strings.Builder
	for _, v := range data {
		idx := 0
		if maxVal > minVal {
			idx = int((v - minVal) / (maxVal - minVal) * float64(len(sparkChars)-1))
		}
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		spark.WriteRune(sparkChars[idx])
	}

	_, _ = fmt.Fprintf(w, "  %s\n", spark.String())
	_, _ = fmt.Fprintf(w, "  Min: %.1f  Avg: %.1f  Max: %.1f\n", minVal, avg, maxVal)

	return nil
}

// RenderHealthIndicators renders traffic-light health indicators.
func RenderHealthIndicators(title string, indicators []metrics.HealthIndicator, w io.Writer) error {
	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", len(title)))

	for _, ind := range indicators {
		dot := "●"
		trendArrow := "→"
		switch ind.Trend {
		case "improving":
			trendArrow = "↑"
		case "declining":
			trendArrow = "↓"
		}

		fmt.Fprintf(w, "  %s %-12s  %-20s  %s %s\n",
			dot,
			ind.Dimension,
			formatValue(ind.Dimension, ind.Value),
			ind.Trend,
			trendArrow,
		)
	}
	return nil
}

func formatValue(dimension string, value float64) string {
	switch dimension {
	case "velocity":
		return fmt.Sprintf("%.1f items/sprint", value)
	case "quality":
		return fmt.Sprintf("%.2f defects/item", value)
	case "review":
		return fmt.Sprintf("%.1f iterations avg", value)
	case "backlog":
		return fmt.Sprintf("%.0f%% ready", value)
	case "flow":
		return fmt.Sprintf("%.1f%% efficiency", value)
	default:
		return fmt.Sprintf("%.2f", value)
	}
}

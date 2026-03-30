package dashboard

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/unbound-force/unbound-force/internal/metrics"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Mx F Dashboard</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 900px; margin: 2rem auto; padding: 0 1rem; }
  h1 { border-bottom: 2px solid #333; padding-bottom: 0.5rem; }
  .metric { display: inline-block; margin: 1rem; padding: 1rem; border: 1px solid #ddd; border-radius: 8px; min-width: 200px; }
  .green { border-left: 4px solid #2da44e; }
  .yellow { border-left: 4px solid #d29922; }
  .red { border-left: 4px solid #cf222e; }
  .value { font-size: 2rem; font-weight: bold; }
  .label { color: #666; font-size: 0.9rem; }
  .trend { font-size: 0.8rem; color: #888; }
</style>
</head>
<body>
<h1>Mx F — Metrics Dashboard</h1>
<p>Generated: {{.Timestamp}}</p>
<h2>Health Overview</h2>
<div>
{{range .Indicators}}
<div class="metric {{.Status}}">
  <div class="label">{{.Dimension}}</div>
  <div class="value">{{.FormattedValue}}</div>
  <div class="trend">{{.Trend}} {{.TrendArrow}}</div>
</div>
{{end}}
</div>
<h2>Key Metrics</h2>
<ul>
  <li>Velocity: {{printf "%.1f" .Snapshot.Velocity}} items/sprint</li>
  <li>Cycle Time: {{printf "%.1f" .Snapshot.CycleTime.Avg}}h avg / {{printf "%.1f" .Snapshot.CycleTime.Median}}h median</li>
  <li>CI Pass Rate: {{printf "%.1f" .Snapshot.CIPassRate}}%</li>
  <li>Flow Efficiency: {{printf "%.1f" .Snapshot.FlowEfficiency}}%</li>
  <li>Review Iterations: {{printf "%.1f" .Snapshot.ReviewIterations}} avg</li>
</ul>
</body>
</html>`

type htmlData struct {
	Timestamp  string
	Snapshot   metrics.MetricsSnapshot
	Indicators []htmlIndicator
}

type htmlIndicator struct {
	Dimension      string
	Status         string
	FormattedValue string
	Trend          string
	TrendArrow     string
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// RenderHTML generates a standalone HTML dashboard file.
func RenderHTML(snapshot metrics.MetricsSnapshot, indicators []metrics.HealthIndicator, outPath string) error {
	tmpl, err := template.New("dashboard").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var htmlInds []htmlIndicator
	for _, ind := range indicators {
		arrow := "→"
		switch ind.Trend {
		case "improving":
			arrow = "↑"
		case "declining":
			arrow = "↓"
		}
		htmlInds = append(htmlInds, htmlIndicator{
			Dimension:      capitalize(ind.Dimension),
			Status:         ind.Status,
			FormattedValue: formatValue(ind.Dimension, ind.Value),
			Trend:          ind.Trend,
			TrendArrow:     arrow,
		})
	}

	data := htmlData{
		Timestamp:  snapshot.Timestamp.Format("2006-01-02 15:04 UTC"),
		Snapshot:   snapshot,
		Indicators: htmlInds,
	}

	outPath = filepath.Clean(outPath)
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create HTML file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return tmpl.Execute(f, data)
}

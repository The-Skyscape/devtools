package charting

import (
	"fmt"
	"html/template"
	"strings"
)

// ChartData represents data for a chart
type ChartData struct {
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
	Title  string    `json:"title"`
	Unit   string    `json:"unit"` // e.g., "$", "%", ""
}

// RenderLineChart generates an SVG line chart from the data
func RenderLineChart(data *ChartData, width, height int) template.HTML {
	if data == nil || len(data.Data) == 0 {
		return renderEmptyState(data.Title)
	}
	
	// Chart dimensions and margins
	marginTop := 20
	marginRight := 20
	marginBottom := 40
	marginLeft := 50
	
	chartWidth := width - marginLeft - marginRight
	chartHeight := height - marginTop - marginBottom
	
	// Find min and max values for scaling
	minVal, maxVal := findMinMax(data.Data)
	
	// Add some padding to the range
	range_ := maxVal - minVal
	if range_ == 0 {
		range_ = 1
	}
	padding := range_ * 0.1
	minVal -= padding
	maxVal += padding
	
	// Generate SVG
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" class="chart-svg">`, width, height))
	
	// Add styles with hover effects
	svg.WriteString(`
	<style>
		.chart-line { fill: none; stroke: rgb(59, 130, 246); stroke-width: 2; transition: stroke-width 0.2s; }
		.chart-area { fill: rgb(59, 130, 246); opacity: 0.1; transition: opacity 0.2s; }
		.chart-point { 
			fill: rgb(59, 130, 246); 
			r: 3;
			transition: r 0.2s, fill 0.2s;
			cursor: pointer;
		}
		.chart-point:hover { 
			r: 5;
			fill: rgb(37, 99, 235);
		}
		.chart-grid { stroke: currentColor; opacity: 0.1; stroke-dasharray: 2,2; }
		.chart-axis { stroke: currentColor; opacity: 0.2; }
		.chart-text { fill: currentColor; opacity: 0.6; font-size: 11px; }
		.chart-title { fill: currentColor; font-size: 14px; font-weight: 600; }
		.chart-svg:hover .chart-line { stroke-width: 3; }
		.chart-svg:hover .chart-area { opacity: 0.15; }
		.chart-tooltip {
			fill: rgb(31, 41, 55);
			stroke: rgb(229, 231, 235);
			stroke-width: 1;
		}
		.chart-tooltip-text {
			fill: white;
			font-size: 12px;
			font-weight: 500;
		}
		.chart-hover-tooltip {
			transition: opacity 0.2s;
		}
		.chart-point-group:hover .chart-hover-tooltip {
			opacity: 1 !important;
		}
	</style>`)
	
	// Draw grid lines
	gridLines := 5
	for i := 0; i <= gridLines; i++ {
		y := marginTop + (chartHeight * i / gridLines)
		svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="chart-grid"/>`,
			marginLeft, y, marginLeft+chartWidth, y))
		
		// Y-axis labels
		value := maxVal - (maxVal-minVal)*float64(i)/float64(gridLines)
		label := formatValue(value, data.Unit)
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="end" class="chart-text">%s</text>`,
			marginLeft-5, y+3, label))
	}
	
	// Draw axes
	svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="chart-axis"/>`,
		marginLeft, marginTop, marginLeft, marginTop+chartHeight))
	svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="chart-axis"/>`,
		marginLeft, marginTop+chartHeight, marginLeft+chartWidth, marginTop+chartHeight))
	
	// Calculate points
	points := make([]point, len(data.Data))
	for i, value := range data.Data {
		var x int
		if len(data.Data) == 1 {
			x = marginLeft + chartWidth/2
		} else {
			x = marginLeft + (chartWidth * i / (len(data.Data) - 1))
		}
		
		var y int
		if maxVal == minVal {
			// All values are the same, place in middle
			y = marginTop + chartHeight/2
		} else {
			y = marginTop + chartHeight - int(((value-minVal)/(maxVal-minVal))*float64(chartHeight))
		}
		points[i] = point{x, y}
	}
	
	// Draw area under the line
	if len(points) > 1 {
		areaPath := fmt.Sprintf("M %d,%d ", points[0].x, marginTop+chartHeight)
		for _, p := range points {
			areaPath += fmt.Sprintf("L %d,%d ", p.x, p.y)
		}
		areaPath += fmt.Sprintf("L %d,%d Z", points[len(points)-1].x, marginTop+chartHeight)
		svg.WriteString(fmt.Sprintf(`<path d="%s" class="chart-area"/>`, areaPath))
	}
	
	// Draw line
	if len(points) > 1 {
		linePath := fmt.Sprintf("M %d,%d ", points[0].x, points[0].y)
		for i := 1; i < len(points); i++ {
			linePath += fmt.Sprintf("L %d,%d ", points[i].x, points[i].y)
		}
		svg.WriteString(fmt.Sprintf(`<path d="%s" class="chart-line"/>`, linePath))
	}
	
	// Draw points with hover tooltips
	for i, p := range points {
		// Create a group for point and its tooltip
		svg.WriteString(fmt.Sprintf(`<g class="chart-point-group">`))
		
		// The actual point
		svg.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" class="chart-point">`, p.x, p.y))
		svg.WriteString(fmt.Sprintf(`<title>%s: %s</title>`, data.Labels[i], formatValue(data.Data[i], data.Unit)))
		svg.WriteString(`</circle>`)
		
		// Enhanced tooltip on hover (positioned above the point)
		tooltipText := fmt.Sprintf("%s: %s", data.Labels[i], formatValue(data.Data[i], data.Unit))
		tooltipWidth := len(tooltipText) * 7 // Approximate width based on text length
		tooltipX := p.x - tooltipWidth/2
		tooltipY := p.y - 25
		
		// Ensure tooltip stays within bounds
		if tooltipX < 5 {
			tooltipX = 5
		}
		if tooltipX+tooltipWidth > width-5 {
			tooltipX = width - tooltipWidth - 5
		}
		if tooltipY < 20 {
			tooltipY = p.y + 25 // Show below if too high
		}
		
		svg.WriteString(fmt.Sprintf(`<g class="chart-hover-tooltip" style="opacity: 0; pointer-events: none;">
			<rect x="%d" y="%d" width="%d" height="20" rx="3" class="chart-tooltip"/>
			<text x="%d" y="%d" text-anchor="middle" class="chart-tooltip-text">%s</text>
		</g>`, tooltipX, tooltipY-15, tooltipWidth, tooltipX+tooltipWidth/2, tooltipY, tooltipText))
		
		svg.WriteString(`</g>`)
	}
	
	// Draw X-axis labels (show every nth label to avoid overlap)
	labelStep := max(1, len(data.Labels)/5)
	for i := 0; i < len(data.Labels); i += labelStep {
		if i < len(points) {
			svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="chart-text">%s</text>`,
				points[i].x, marginTop+chartHeight+15, data.Labels[i]))
		}
	}
	
	// Draw title
	if data.Title != "" {
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="chart-title">%s</text>`,
			width/2, 15, data.Title))
	}
	
	svg.WriteString(`</svg>`)
	
	return template.HTML(svg.String())
}

// RenderEmptyState generates a placeholder for empty charts
func renderEmptyState(title string) template.HTML {
	return template.HTML(fmt.Sprintf(`
	<div class="flex flex-col items-center justify-center h-48 text-base-content/60">
		<svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mb-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
		</svg>
		<p class="text-sm">No data available for %s</p>
	</div>`, title))
}

type point struct {
	x, y int
}

func findMinMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	// If all values are the same, add some range
	if min == max {
		if min == 0 {
			return -1, 1
		}
		return min * 0.9, max * 1.1
	}
	
	return min, max
}

func formatValue(value float64, unit string) string {
	if unit == "$" {
		if value >= 1000000 {
			return fmt.Sprintf("$%.1fM", value/1000000)
		} else if value >= 1000 {
			return fmt.Sprintf("$%.1fk", value/1000)
		}
		return fmt.Sprintf("$%.0f", value)
	} else if unit == "%" {
		return fmt.Sprintf("%.0f%%", value)
	}
	
	// No unit, just format the number
	if value >= 1000000 {
		return fmt.Sprintf("%.1fM", value/1000000)
	} else if value >= 1000 {
		return fmt.Sprintf("%.1fk", value/1000)
	}
	return fmt.Sprintf("%.0f", value)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RenderSparkline generates a minimal inline sparkline
func RenderSparkline(data []float64, width, height int) template.HTML {
	if len(data) == 0 {
		return template.HTML("")
	}
	
	minVal, maxVal := findMinMax(data)
	range_ := maxVal - minVal
	if range_ == 0 {
		range_ = 1
	}
	
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" style="vertical-align: middle;">`, width, height))
	
	// Add hover effect style for sparkline
	svg.WriteString(`<style>
		.sparkline-path { transition: stroke-width 0.2s, stroke 0.2s; }
		.sparkline-path:hover { stroke-width: 3; stroke: rgb(37, 99, 235); }
	</style>`)
	
	// Generate path
	path := "M"
	for i, value := range data {
		var x int
		if len(data) == 1 {
			x = width / 2
		} else {
			x = (width * i) / (len(data) - 1)
		}
		
		var y int
		if range_ == 0 {
			y = height / 2
		} else {
			y = height - int(((value-minVal)/range_)*float64(height))
		}
		
		if i == 0 {
			path += fmt.Sprintf(" %d,%d", x, y)
		} else {
			path += fmt.Sprintf(" L %d,%d", x, y)
		}
	}
	
	svg.WriteString(fmt.Sprintf(`<path d="%s" fill="none" stroke="rgb(59, 130, 246)" stroke-width="2" class="sparkline-path"/>`, path))
	svg.WriteString(`</svg>`)
	
	return template.HTML(svg.String())
}

// PlaceholderChart generates a placeholder chart with a message
func PlaceholderChart(title, message string) template.HTML {
	return template.HTML(fmt.Sprintf(`
	<div class="flex flex-col items-center justify-center h-64 bg-base-200/30 rounded-lg border border-base-300">
		<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
		</svg>
		<p class="text-lg font-medium text-base-content/70">%s</p>
		<p class="text-sm text-base-content/50 mt-1">%s</p>
	</div>`, title, message))
}
package containers

import (
	"context"
	"sync"
	"time"
)

// Collector manages resource collection and history
type Collector struct {
	monitor   *Monitor
	history   []SystemStats
	maxHistory int
	mu        sync.RWMutex
	interval  time.Duration
	cancel    context.CancelFunc
}

// NewCollector creates a new resource collector
func NewCollector(includeContainers bool, maxHistory int) *Collector {
	return &Collector{
		monitor:    NewMonitor(includeContainers),
		history:    make([]SystemStats, 0, maxHistory),
		maxHistory: maxHistory,
		interval:   5 * time.Second,
	}
}

// Start begins collecting system statistics
func (c *Collector) Start() {
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		return // Already running
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.mu.Unlock()

	go c.collect(ctx)
}

// Stop stops collecting statistics
func (c *Collector) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.mu.Unlock()
}

// SetInterval updates the collection interval
func (c *Collector) SetInterval(interval time.Duration) {
	c.mu.Lock()
	c.interval = interval
	c.mu.Unlock()
}

// collect continuously collects system statistics
func (c *Collector) collect(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect initial stats
	c.collectOnce()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collectOnce()
		}
	}
}

// collectOnce collects system statistics once
func (c *Collector) collectOnce() {
	stats, err := c.monitor.GetSystemStats()
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.history = append(c.history, *stats)
	
	// Trim history if needed
	if len(c.history) > c.maxHistory {
		c.history = c.history[len(c.history)-c.maxHistory:]
	}
}

// GetCurrent returns the most recent system statistics
func (c *Collector) GetCurrent() (*SystemStats, error) {
	// Try to get from history first
	c.mu.RLock()
	if len(c.history) > 0 {
		stats := c.history[len(c.history)-1]
		c.mu.RUnlock()
		return &stats, nil
	}
	c.mu.RUnlock()

	// If no history, collect now
	return c.monitor.GetSystemStats()
}

// GetHistory returns the collected statistics history
func (c *Collector) GetHistory() []SystemStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := make([]SystemStats, len(c.history))
	copy(result, c.history)
	return result
}

// GetHistorySince returns statistics collected since the given time
func (c *Collector) GetHistorySince(since time.Time) []SystemStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []SystemStats
	for _, stat := range c.history {
		if stat.Timestamp.After(since) {
			result = append(result, stat)
		}
	}
	return result
}

// GetAverages calculates average statistics over a time period
func (c *Collector) GetAverages(duration time.Duration) *SystemStats {
	since := time.Now().Add(-duration)
	stats := c.GetHistorySince(since)
	
	if len(stats) == 0 {
		return nil
	}

	avg := &SystemStats{
		Timestamp: time.Now(),
	}

	// Calculate averages
	var cpuTotal, memTotal, diskTotal float64
	var loadTotal1, loadTotal5, loadTotal15 float64
	
	for _, s := range stats {
		cpuTotal += s.CPU.UsagePercent
		memTotal += s.Memory.UsedPercent
		diskTotal += s.Disk.UsedPercent
		loadTotal1 += s.LoadAverage.Load1
		loadTotal5 += s.LoadAverage.Load5
		loadTotal15 += s.LoadAverage.Load15
	}

	count := float64(len(stats))
	avg.CPU.UsagePercent = cpuTotal / count
	avg.Memory.UsedPercent = memTotal / count
	avg.Disk.UsedPercent = diskTotal / count
	avg.LoadAverage.Load1 = loadTotal1 / count
	avg.LoadAverage.Load5 = loadTotal5 / count
	avg.LoadAverage.Load15 = loadTotal15 / count

	// Use latest values for totals
	latest := stats[len(stats)-1]
	avg.CPU.Cores = latest.CPU.Cores
	avg.Memory.Total = latest.Memory.Total
	avg.Memory.Available = latest.Memory.Available
	avg.Disk.Total = latest.Disk.Total
	avg.Disk.Free = latest.Disk.Free

	return avg
}

// ClearHistory clears the collected statistics history
func (c *Collector) ClearHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = c.history[:0]
}

// ResourceAlert represents a resource usage alert
type ResourceAlert struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// CheckAlerts checks for resource usage alerts
func (c *Collector) CheckAlerts() []ResourceAlert {
	var alerts []ResourceAlert

	stats, err := c.GetCurrent()
	if err != nil {
		return alerts
	}

	// Check CPU usage
	if stats.CPU.UsagePercent > 90 {
		alerts = append(alerts, ResourceAlert{
			Type:      "cpu",
			Severity:  "critical",
			Message:   "CPU usage is critically high",
			Value:     stats.CPU.UsagePercent,
			Threshold: 90,
			Timestamp: stats.Timestamp,
		})
	} else if stats.CPU.UsagePercent > 70 {
		alerts = append(alerts, ResourceAlert{
			Type:      "cpu",
			Severity:  "warning",
			Message:   "CPU usage is high",
			Value:     stats.CPU.UsagePercent,
			Threshold: 70,
			Timestamp: stats.Timestamp,
		})
	}

	// Check memory usage
	if stats.Memory.UsedPercent > 90 {
		alerts = append(alerts, ResourceAlert{
			Type:      "memory",
			Severity:  "critical",
			Message:   "Memory usage is critically high",
			Value:     stats.Memory.UsedPercent,
			Threshold: 90,
			Timestamp: stats.Timestamp,
		})
	} else if stats.Memory.UsedPercent > 80 {
		alerts = append(alerts, ResourceAlert{
			Type:      "memory",
			Severity:  "warning",
			Message:   "Memory usage is high",
			Value:     stats.Memory.UsedPercent,
			Threshold: 80,
			Timestamp: stats.Timestamp,
		})
	}

	// Check disk usage
	if stats.Disk.UsedPercent > 95 {
		alerts = append(alerts, ResourceAlert{
			Type:      "disk",
			Severity:  "critical",
			Message:   "Disk usage is critically high",
			Value:     stats.Disk.UsedPercent,
			Threshold: 95,
			Timestamp: stats.Timestamp,
		})
	} else if stats.Disk.UsedPercent > 85 {
		alerts = append(alerts, ResourceAlert{
			Type:      "disk",
			Severity:  "warning",
			Message:   "Disk usage is high",
			Value:     stats.Disk.UsedPercent,
			Threshold: 85,
			Timestamp: stats.Timestamp,
		})
	}

	// Check load average (relative to CPU cores)
	if stats.CPU.Cores > 0 {
		loadPerCore := stats.LoadAverage.Load5 / float64(stats.CPU.Cores)
		if loadPerCore > 2.0 {
			alerts = append(alerts, ResourceAlert{
				Type:      "load",
				Severity:  "warning",
				Message:   "System load is high",
				Value:     stats.LoadAverage.Load5,
				Threshold: float64(stats.CPU.Cores) * 2.0,
				Timestamp: stats.Timestamp,
			})
		}
	}

	return alerts
}
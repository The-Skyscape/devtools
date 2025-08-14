package monitoring

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SystemStats represents system resource usage
type SystemStats struct {
	Timestamp    time.Time         `json:"timestamp"`
	CPU          CPUStats          `json:"cpu"`
	Memory       MemoryStats       `json:"memory"`
	Disk         DiskStats         `json:"disk"`
	LoadAverage  LoadStats         `json:"load_average"`
	Processes    ProcessStats      `json:"processes"`
	Network      NetworkStats      `json:"network"`
	Containers   []ContainerStats  `json:"containers,omitempty"`
}

// CPUStats represents CPU usage information
type CPUStats struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
	UserPercent  float64 `json:"user_percent"`
	SystemPercent float64 `json:"system_percent"`
	IdlePercent  float64 `json:"idle_percent"`
}

// MemoryStats represents memory usage information
type MemoryStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
	Cached      uint64  `json:"cached"`
	Buffers     uint64  `json:"buffers"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapFree    uint64  `json:"swap_free"`
}

// DiskStats represents disk usage information
type DiskStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	Mountpoint  string  `json:"mountpoint"`
}

// LoadStats represents system load averages
type LoadStats struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// ProcessStats represents process information
type ProcessStats struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Sleeping int `json:"sleeping"`
	Stopped int `json:"stopped"`
	Zombie  int `json:"zombie"`
}

// NetworkStats represents network statistics
type NetworkStats struct {
	BytesRecv   uint64 `json:"bytes_recv"`
	BytesSent   uint64 `json:"bytes_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	PacketsSent uint64 `json:"packets_sent"`
}

// ContainerStats represents Docker container statistics
type ContainerStats struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	CPUPercent   float64 `json:"cpu_percent"`
	MemUsage     uint64  `json:"mem_usage"`
	MemLimit     uint64  `json:"mem_limit"`
	MemPercent   float64 `json:"mem_percent"`
	NetIO        string  `json:"net_io"`
	BlockIO      string  `json:"block_io"`
}

// Monitor provides system monitoring capabilities
type Monitor struct {
	includeContainers bool
}

// NewMonitor creates a new system monitor
func NewMonitor(includeContainers bool) *Monitor {
	return &Monitor{
		includeContainers: includeContainers,
	}
}

// GetSystemStats retrieves current system statistics
func (m *Monitor) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{
		Timestamp: time.Now(),
	}

	// Get CPU stats
	cpu, err := m.getCPUStats()
	if err == nil {
		stats.CPU = *cpu
	}

	// Get memory stats
	mem, err := m.getMemoryStats()
	if err == nil {
		stats.Memory = *mem
	}

	// Get disk stats
	disk, err := m.getDiskStats("/")
	if err == nil {
		stats.Disk = *disk
	}

	// Get load average
	load, err := m.getLoadStats()
	if err == nil {
		stats.LoadAverage = *load
	}

	// Get process stats
	procs, err := m.getProcessStats()
	if err == nil {
		stats.Processes = *procs
	}

	// Get network stats
	net, err := m.getNetworkStats()
	if err == nil {
		stats.Network = *net
	}

	// Get container stats if requested
	if m.includeContainers {
		containers, err := m.getContainerStats()
		if err == nil {
			stats.Containers = containers
		}
	}

	return stats, nil
}

// getCPUStats retrieves CPU usage statistics
func (m *Monitor) getCPUStats() (*CPUStats, error) {
	stats := &CPUStats{
		Cores: runtime.NumCPU(),
	}

	// Read /proc/stat for CPU usage
	file, err := os.Open("/proc/stat")
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				continue
			}

			var user, nice, system, idle, iowait, irq, softirq uint64
			user, _ = strconv.ParseUint(fields[1], 10, 64)
			nice, _ = strconv.ParseUint(fields[2], 10, 64)
			system, _ = strconv.ParseUint(fields[3], 10, 64)
			idle, _ = strconv.ParseUint(fields[4], 10, 64)
			iowait, _ = strconv.ParseUint(fields[5], 10, 64)
			irq, _ = strconv.ParseUint(fields[6], 10, 64)
			softirq, _ = strconv.ParseUint(fields[7], 10, 64)

			total := user + nice + system + idle + iowait + irq + softirq
			if total > 0 {
				stats.UserPercent = float64(user+nice) * 100.0 / float64(total)
				stats.SystemPercent = float64(system+irq+softirq) * 100.0 / float64(total)
				stats.IdlePercent = float64(idle+iowait) * 100.0 / float64(total)
				stats.UsagePercent = 100.0 - stats.IdlePercent
			}
			break
		}
	}

	return stats, nil
}

// getMemoryStats retrieves memory usage statistics
func (m *Monitor) getMemoryStats() (*MemoryStats, error) {
	stats := &MemoryStats{}

	// Read /proc/meminfo
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseUint(fields[1], 10, 64)
		value *= 1024 // Convert from KB to bytes

		switch fields[0] {
		case "MemTotal:":
			stats.Total = value
		case "MemFree:":
			stats.Free = value
		case "MemAvailable:":
			stats.Available = value
		case "Cached:":
			stats.Cached = value
		case "Buffers:":
			stats.Buffers = value
		case "SwapTotal:":
			stats.SwapTotal = value
		case "SwapFree:":
			stats.SwapFree = value
		}
	}

	stats.Used = stats.Total - stats.Available
	if stats.Total > 0 {
		stats.UsedPercent = float64(stats.Used) * 100.0 / float64(stats.Total)
	}
	stats.SwapUsed = stats.SwapTotal - stats.SwapFree

	return stats, nil
}

// getDiskStats retrieves disk usage statistics for a given path
func (m *Monitor) getDiskStats(path string) (*DiskStats, error) {
	stats := &DiskStats{
		Mountpoint: path,
	}

	// Use df command to get disk usage
	cmd := exec.Command("df", "-B1", path)
	output, err := cmd.Output()
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return stats, fmt.Errorf("unexpected df output")
	}

	// Parse the second line (first line is header)
	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return stats, fmt.Errorf("unexpected df output format")
	}

	stats.Total, _ = strconv.ParseUint(fields[1], 10, 64)
	stats.Used, _ = strconv.ParseUint(fields[2], 10, 64)
	stats.Free, _ = strconv.ParseUint(fields[3], 10, 64)
	
	percentStr := strings.TrimSuffix(fields[4], "%")
	percent, _ := strconv.ParseFloat(percentStr, 64)
	stats.UsedPercent = percent

	return stats, nil
}

// getLoadStats retrieves system load averages
func (m *Monitor) getLoadStats() (*LoadStats, error) {
	stats := &LoadStats{}

	// Read /proc/loadavg
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return stats, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return stats, fmt.Errorf("unexpected loadavg format")
	}

	stats.Load1, _ = strconv.ParseFloat(fields[0], 64)
	stats.Load5, _ = strconv.ParseFloat(fields[1], 64)
	stats.Load15, _ = strconv.ParseFloat(fields[2], 64)

	return stats, nil
}

// getProcessStats retrieves process statistics
func (m *Monitor) getProcessStats() (*ProcessStats, error) {
	stats := &ProcessStats{}

	// Read /proc/stat for process counts
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "procs_running") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				stats.Running, _ = strconv.Atoi(fields[1])
			}
		}
	}

	// Count total processes from /proc
	entries, err := os.ReadDir("/proc")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				// Check if directory name is a PID (all digits)
				if _, err := strconv.Atoi(entry.Name()); err == nil {
					stats.Total++
				}
			}
		}
	}

	return stats, nil
}

// getNetworkStats retrieves network statistics
func (m *Monitor) getNetworkStats() (*NetworkStats, error) {
	stats := &NetworkStats{}

	// Read /proc/net/dev
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Skip header lines and loopback
		if strings.Contains(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "lo:") {
			fields := strings.Fields(line)
			if len(fields) >= 17 {
				// Accumulate stats from all interfaces except loopback
				recv, _ := strconv.ParseUint(fields[1], 10, 64)
				sent, _ := strconv.ParseUint(fields[9], 10, 64)
				recvPkts, _ := strconv.ParseUint(fields[2], 10, 64)
				sentPkts, _ := strconv.ParseUint(fields[10], 10, 64)
				
				stats.BytesRecv += recv
				stats.BytesSent += sent
				stats.PacketsRecv += recvPkts
				stats.PacketsSent += sentPkts
			}
		}
	}

	return stats, nil
}

// getContainerStats retrieves Docker container statistics
func (m *Monitor) getContainerStats() ([]ContainerStats, error) {
	var containers []ContainerStats

	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return containers, nil // Docker not installed
	}

	// Get container stats using docker stats
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", 
		"{{.Container}}|{{.Name}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}")
	
	output, err := cmd.Output()
	if err != nil {
		return containers, nil // Docker might not be running
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) >= 7 {
			container := ContainerStats{
				ID:      fields[0],
				Name:    fields[1],
				Status:  "running", // All containers from stats are running
				NetIO:   fields[5],
				BlockIO: fields[6],
			}

			// Parse CPU percentage
			cpuStr := strings.TrimSuffix(fields[2], "%")
			container.CPUPercent, _ = strconv.ParseFloat(cpuStr, 64)

			// Parse memory usage (format: "100MiB / 1GiB")
			memParts := strings.Split(fields[3], " / ")
			if len(memParts) == 2 {
				container.MemUsage = parseMemorySize(strings.TrimSpace(memParts[0]))
				container.MemLimit = parseMemorySize(strings.TrimSpace(memParts[1]))
			}

			// Parse memory percentage
			memPercentStr := strings.TrimSuffix(fields[4], "%")
			container.MemPercent, _ = strconv.ParseFloat(memPercentStr, 64)

			containers = append(containers, container)
		}
	}

	return containers, nil
}

// parseMemorySize converts memory size strings (e.g., "100MiB", "1.5GiB") to bytes
func parseMemorySize(size string) uint64 {
	size = strings.TrimSpace(size)
	
	multipliers := map[string]uint64{
		"B":   1,
		"KiB": 1024,
		"MiB": 1024 * 1024,
		"GiB": 1024 * 1024 * 1024,
		"KB":  1000,
		"MB":  1000 * 1000,
		"GB":  1000 * 1000 * 1000,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(size, suffix) {
			numStr := strings.TrimSuffix(size, suffix)
			if val, err := strconv.ParseFloat(numStr, 64); err == nil {
				return uint64(val * float64(multiplier))
			}
		}
	}

	// Try to parse as plain number
	if val, err := strconv.ParseUint(size, 10, 64); err == nil {
		return val
	}

	return 0
}

// StreamStats continuously monitors and sends system statistics
func (m *Monitor) StreamStats(ctx context.Context, interval time.Duration, ch chan<- *SystemStats) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send initial stats immediately
	if stats, err := m.GetSystemStats(); err == nil {
		select {
		case ch <- stats:
		case <-ctx.Done():
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if stats, err := m.GetSystemStats(); err == nil {
				select {
				case ch <- stats:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// FormatBytes formats bytes into human-readable string
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatPercent formats a percentage value
func FormatPercent(percent float64) string {
	return fmt.Sprintf("%.1f%%", percent)
}

// GetTopProcesses returns the top N processes by CPU or memory usage
func (m *Monitor) GetTopProcesses(count int, sortBy string) ([]ProcessInfo, error) {
	var processes []ProcessInfo

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return processes, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is a PID
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		proc, err := m.getProcessInfo(pid)
		if err != nil {
			continue
		}

		processes = append(processes, *proc)
	}

	// Sort processes
	if sortBy == "memory" {
		// Sort by memory usage
		for i := 0; i < len(processes)-1; i++ {
			for j := i + 1; j < len(processes); j++ {
				if processes[j].MemoryKB > processes[i].MemoryKB {
					processes[i], processes[j] = processes[j], processes[i]
				}
			}
		}
	} else {
		// Sort by CPU usage (default)
		for i := 0; i < len(processes)-1; i++ {
			for j := i + 1; j < len(processes); j++ {
				if processes[j].CPUPercent > processes[i].CPUPercent {
					processes[i], processes[j] = processes[j], processes[i]
				}
			}
		}
	}

	// Return top N
	if count > 0 && count < len(processes) {
		processes = processes[:count]
	}

	return processes, nil
}

// ProcessInfo represents information about a single process
type ProcessInfo struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	User       string  `json:"user"`
	CPUPercent float64 `json:"cpu_percent"`
	MemoryKB   uint64  `json:"memory_kb"`
	State      string  `json:"state"`
	Command    string  `json:"command"`
}

// getProcessInfo retrieves information about a specific process
func (m *Monitor) getProcessInfo(pid int) (*ProcessInfo, error) {
	info := &ProcessInfo{
		PID: pid,
	}

	// Read process status
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		return info, err
	}

	lines := strings.Split(string(statusData), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "Name:":
			info.Name = fields[1]
		case "State:":
			info.State = fields[1]
		case "VmRSS:":
			if len(fields) >= 2 {
				info.MemoryKB, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	// Read command line
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlineData, err := os.ReadFile(cmdlinePath)
	if err == nil {
		// Replace null bytes with spaces
		info.Command = strings.ReplaceAll(string(bytes.ReplaceAll(cmdlineData, []byte{0}, []byte{' '})), "  ", " ")
		info.Command = strings.TrimSpace(info.Command)
		if info.Command == "" {
			info.Command = fmt.Sprintf("[%s]", info.Name)
		}
	}

	return info, nil
}
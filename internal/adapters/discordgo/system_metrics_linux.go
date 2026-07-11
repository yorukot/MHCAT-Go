//go:build linux

package discordgo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	runtimemetrics "runtime/metrics"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const bytesPerMiB = 1024 * 1024

type linuxSystemMetricsSampler struct {
	CPUInterval time.Duration
}

func defaultSystemMetricsSampler() SystemMetricsSampler {
	return linuxSystemMetricsSampler{CPUInterval: time.Second}
}

func (s linuxSystemMetricsSampler) Sample(ctx context.Context) (SystemMetrics, error) {
	model, err := linuxCPUModel()
	if err != nil {
		return SystemMetrics{}, err
	}
	usage, err := linuxCPUUsage(ctx, s.CPUInterval)
	if err != nil {
		return SystemMetrics{}, err
	}
	hostUsed, hostTotal, err := linuxHostMemory()
	if err != nil {
		return SystemMetrics{}, err
	}
	processHeap, processRSS, err := s.SampleProcess(ctx)
	if err != nil {
		return SystemMetrics{}, err
	}
	return SystemMetrics{
		CPUModel:          model,
		CPUUsagePercent:   usage,
		HostMemoryUsedMB:  hostUsed,
		HostMemoryTotalMB: hostTotal,
		ProcessHeapMB:     processHeap,
		ProcessRSSMB:      processRSS,
	}, nil
}

func (linuxSystemMetricsSampler) SampleProcess(ctx context.Context) (heapMB int64, rssMB int64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}
	return linuxProcessMemory()
}

func linuxCPUModel() (string, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", fmt.Errorf("open cpu info: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), ":")
		if found && strings.TrimSpace(key) == "model name" {
			return strings.TrimSpace(value), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read cpu info: %w", err)
	}
	return "unknown", nil
}

func linuxCPUUsage(ctx context.Context, interval time.Duration) (float64, error) {
	firstIdle, firstTotal, err := linuxCPUTimes()
	if err != nil {
		return 0, err
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-timer.C:
	}
	secondIdle, secondTotal, err := linuxCPUTimes()
	if err != nil {
		return 0, err
	}
	totalDelta := secondTotal - firstTotal
	if totalDelta == 0 {
		return 0, nil
	}
	idleDelta := secondIdle - firstIdle
	return (1 - float64(idleDelta)/float64(totalDelta)) * 100, nil
}

func linuxCPUTimes() (idle uint64, total uint64, err error) {
	payload, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, fmt.Errorf("read cpu stats: %w", err)
	}
	line, _, _ := strings.Cut(string(payload), "\n")
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, fmt.Errorf("unexpected cpu stats line %q", line)
	}
	values := make([]uint64, 0, len(fields)-1)
	for _, field := range fields[1:] {
		value, parseErr := strconv.ParseUint(field, 10, 64)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("parse cpu stat %q: %w", field, parseErr)
		}
		values = append(values, value)
		total += value
	}
	idle = values[3]
	return idle, total, nil
}

func linuxHostMemory() (usedMB int64, totalMB int64, err error) {
	var info unix.Sysinfo_t
	if err := unix.Sysinfo(&info); err != nil {
		return 0, 0, fmt.Errorf("read host memory: %w", err)
	}
	unit := uint64(info.Unit)
	if unit == 0 {
		unit = 1
	}
	total := uint64(info.Totalram) * unit
	free := uint64(info.Freeram) * unit
	return roundedMiB(total - free), roundedMiB(total), nil
}

func linuxProcessMemory() (heapMB int64, rssMB int64, err error) {
	samples := [...]runtimemetrics.Sample{
		{Name: "/memory/classes/heap/objects:bytes"},
		{Name: "/memory/classes/heap/unused:bytes"},
		{Name: "/memory/classes/heap/free:bytes"},
		{Name: "/memory/classes/heap/released:bytes"},
	}
	runtimemetrics.Read(samples[:])
	var heapBytes uint64
	for _, sample := range samples {
		if sample.Value.Kind() != runtimemetrics.KindUint64 {
			return 0, 0, fmt.Errorf("unexpected runtime metric kind for %s", sample.Name)
		}
		heapBytes += sample.Value.Uint64()
	}
	payload, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, 0, fmt.Errorf("read process memory: %w", err)
	}
	fields := strings.Fields(string(payload))
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected process memory %q", strings.TrimSpace(string(payload)))
	}
	residentPages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse resident pages: %w", err)
	}
	return roundedMiB(heapBytes), roundedMiB(residentPages * uint64(os.Getpagesize())), nil
}

func roundedMiB(bytes uint64) int64 {
	return int64((bytes + bytesPerMiB/2) / bytesPerMiB)
}

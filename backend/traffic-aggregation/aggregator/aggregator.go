package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"visual-network/traffic-aggregation/ebpf"
	"visual-network/traffic-aggregation/storage"
	"visual-network/traffic-aggregation/types"
)

const (
	FlowTimeout         = 10 * time.Second
	AggregationInterval = 1 * time.Second
	CleanupInterval     = 5 * time.Second
)

// Aggregator processes eBPF data and prepares topology updates
type Aggregator struct {
	monitor      *ebpf.Monitor
	store        *storage.Store
	flows        map[types.FlowKey]*types.Flow
	flowsMutex   sync.RWMutex
	ipCache      map[string]*types.IPInfo
	ipCacheMutex sync.RWMutex
	updateChan   chan types.TopologyUpdate
	metricsChan  chan types.MetricsUpdate
	stopChan     chan struct{}
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewAggregator creates a new aggregator with default storage
func NewAggregator(monitor *ebpf.Monitor) *Aggregator {
	return NewAggregatorWithStorage(monitor, "./traffic-aggregation/storage/jsonDumps/", 100)
}

// NewAggregatorWithStorage creates a new aggregator with custom storage configuration
func NewAggregatorWithStorage(monitor *ebpf.Monitor, dataDir string, maxFiles int) *Aggregator {
	ctx, cancel := context.WithCancel(context.Background())

	// Create storage
	store, err := storage.NewStore(dataDir, maxFiles)
	if err != nil {
		log.Printf("Warning: failed to initialize storage: %v", err)
	}

	a := &Aggregator{
		monitor:     monitor,
		store:       store,
		flows:       make(map[types.FlowKey]*types.Flow),
		ipCache:     make(map[string]*types.IPInfo),
		updateChan:  make(chan types.TopologyUpdate, 10),
		metricsChan: make(chan types.MetricsUpdate, 10),
		stopChan:    make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background workers
	go a.processEvents()
	go a.aggregateLoop()
	go a.cleanupLoop()

	log.Println("Aggregator started")
	return a
}

// processEvents handles new connection events from eBPF
func (a *Aggregator) processEvents() {
	for {
		select {
		case <-a.stopChan:
			return
		case event := <-a.monitor.EventChannel():
			a.handleNewConnection(event)
		}
	}
}

// handleNewConnection processes a new connection event
func (a *Aggregator) handleNewConnection(event types.ConnEvent) {
	key := types.FlowKey{
		SrcIP:    event.SrcIP,
		DstIP:    event.DstIP,
		SrcPort:  event.SrcPort,
		DstPort:  event.DstPort,
		Protocol: event.Protocol,
	}

	a.flowsMutex.Lock()
	defer a.flowsMutex.Unlock()

	// Check if we already have this flow
	if _, exists := a.flows[key]; exists {
		return
	}

	// Create new flow
	srcIP := ebpf.Uint32ToIP(event.SrcIP).String()
	dstIP := ebpf.Uint32ToIP(event.DstIP).String()

	flow := &types.Flow{
		Key:          key,
		SrcIPStr:     srcIP,
		DstIPStr:     dstIP,
		ProtocolName: types.ProtocolToString(event.Protocol),
		L7ProtoName:  types.L7ProtocolToString(event.L7Protocol),
		Active:       true,
		LastUpdate:   time.Now(),
	}

	// Resolve hostnames asynchronously
	go a.resolveIP(srcIP)
	go a.resolveIP(dstIP)

	a.flows[key] = flow

	log.Printf("New flow: %s:%d -> %s:%d (%s)",
		srcIP, event.SrcPort, dstIP, event.DstPort, flow.ProtocolName)
}

// aggregateLoop periodically reads eBPF maps and sends updates
func (a *Aggregator) aggregateLoop() {
	ticker := time.NewTicker(AggregationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.aggregate()
		}
	}
}

// aggregate reads current state from eBPF and prepares updates
func (a *Aggregator) aggregate() {
	// Read all flows from eBPF
	ebpfFlows, err := a.monitor.GetAllFlows()
	if err != nil {
		log.Printf("Error reading flows: %v", err)
		return
	}

	a.flowsMutex.Lock()

	var newFlows []types.Flow
	var updatedFlows []types.Flow
	var removedKeys []types.FlowKey

	// Update existing flows and find new ones
	for key, stats := range ebpfFlows {
		flow, exists := a.flows[key]

		if !exists {
			// This shouldn't happen often (events should catch new flows)
			// but handle it just in case
			srcIP := ebpf.Uint32ToIP(key.SrcIP).String()
			dstIP := ebpf.Uint32ToIP(key.DstIP).String()

			flow = &types.Flow{
				Key:          key,
				SrcIPStr:     srcIP,
				DstIPStr:     dstIP,
				ProtocolName: types.ProtocolToString(key.Protocol),
				L7ProtoName:  types.L7ProtocolToString(stats.L7Protocol),
				Active:       true,
				LastUpdate:   time.Now(),
			}
			a.flows[key] = flow

			go a.resolveIP(srcIP)
			go a.resolveIP(dstIP)
		}

		// Update flow stats
		flow.Stats = stats
		flow.LastUpdate = time.Now()

		// Calculate average latency
		if stats.LatencyCount > 0 {
			flow.AvgLatencyMs = float64(stats.LatencySum) / float64(stats.LatencyCount) / 1000.0
		} else if stats.LastSeen > stats.FirstSeen {
			deltaNs := stats.LastSeen - stats.FirstSeen
			if stats.Packets > 1 {
				flow.AvgLatencyMs = float64(deltaNs) / float64(stats.Packets-1) / 1_000_000.0
			} else {
				flow.AvgLatencyMs = float64(deltaNs) / 1_000_000.0
			}
		} else {
			flow.AvgLatencyMs = 0
		}

		// Get resolved names
		flow.SrcName = a.getResolvedName(flow.SrcIPStr)
		flow.DstName = a.getResolvedName(flow.DstIPStr)

		if !exists {
			newFlows = append(newFlows, *flow)
		} else {
			updatedFlows = append(updatedFlows, *flow)
		}
	}

	// Find removed flows (exist in our map but not in eBPF)
	for key, flow := range a.flows {
		if _, exists := ebpfFlows[key]; !exists {
			removedKeys = append(removedKeys, key)
			delete(a.flows, key)
			flow.Active = false
		}
	}

	a.flowsMutex.Unlock()

	// Send topology update
	if len(newFlows) > 0 || len(updatedFlows) > 0 || len(removedKeys) > 0 {
		update := types.TopologyUpdate{
			Timestamp: time.Now(),
			NewFlows:  newFlows,
			Updated:   updatedFlows,
			Removed:   removedKeys,
			Summary:   a.calculateSummary(),
		}

		// Save to storage
		if a.store != nil {
			if err := a.store.SaveTopologyUpdate(update); err != nil {
				log.Printf("Error saving topology update: %v", err)
			}
		}

		select {
		case a.updateChan <- update:
		default:
			log.Println("Update channel full, dropping update")
		}
	}

	// Send metrics update (top latency flows)
	metrics := a.calculateMetrics()

	// Save metrics to storage
	if a.store != nil {
		if err := a.store.SaveMetricsUpdate(metrics); err != nil {
			log.Printf("Error saving metrics update: %v", err)
		}
	}

	select {
	case a.metricsChan <- metrics:
	default:
		log.Println("Metrics channel full, dropping metrics")
	}
}

// cleanupLoop periodically removes expired flows
func (a *Aggregator) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			deleted, err := a.monitor.CleanExpiredFlows(FlowTimeout)
			if err != nil {
				log.Printf("Error cleaning flows: %v", err)
			} else if deleted > 0 {
				log.Printf("Cleaned %d expired flows", deleted)
			}
		}
	}
}

// calculateSummary generates aggregate statistics
func (a *Aggregator) calculateSummary() types.TopologySummary {
	a.flowsMutex.RLock()
	defer a.flowsMutex.RUnlock()
	return a.calculateSummaryLocked()
}


func (a *Aggregator) calculateSummaryLocked() types.TopologySummary {
	summary := types.TopologySummary{
		TotalFlows: len(a.flows),
	}

	uniqueIPs := make(map[string]bool)
	var totalLatency float64
	var latencyCount int

	for _, flow := range a.flows {
		summary.TotalPackets += flow.Stats.Packets
		summary.TotalBytes += flow.Stats.Bytes

		uniqueIPs[flow.SrcIPStr] = true
		uniqueIPs[flow.DstIPStr] = true

		if flow.AvgLatencyMs > 0 {
			totalLatency += flow.AvgLatencyMs
			latencyCount++
		}
	}

	summary.UniqueIPs = len(uniqueIPs)
	if latencyCount > 0 {
		summary.AvgLatencyMs = totalLatency / float64(latencyCount)
	}

	return summary
}

// calculateMetrics generates top N latency flows
func (a *Aggregator) calculateMetrics() types.MetricsUpdate {
	a.flowsMutex.RLock()
	defer a.flowsMutex.RUnlock()
	return a.calculateMetricsLocked()
}


func (a *Aggregator) calculateMetricsLocked() types.MetricsUpdate {
	// Collect all flows with latency data
	var latencyFlows []types.LatencyInfo
	for _, flow := range a.flows {
		if flow.AvgLatencyMs > 0 {
			latencyFlows = append(latencyFlows, types.LatencyInfo{
				SrcIP:        flow.SrcIPStr,
				DstIP:        flow.DstIPStr,
				SrcName:      flow.SrcName,
				DstName:      flow.DstName,
				AvgLatencyMs: flow.AvgLatencyMs,
				Protocol:     flow.ProtocolName,
			})
		}
	}

	// Sort by latency (highest first)
	sort.Slice(latencyFlows, func(i, j int) bool {
		return latencyFlows[i].AvgLatencyMs > latencyFlows[j].AvgLatencyMs
	})

	// Get top 5 by source and destination
	metrics := types.MetricsUpdate{
		Timestamp:     time.Now(),
		TopLatencySrc: make([]types.LatencyInfo, 0, 5),
		TopLatencyDst: make([]types.LatencyInfo, 0, 5),
	}

	// Top 5 overall (grouped by source or destination as needed)
	seen := make(map[string]bool)
	for _, info := range latencyFlows {
		// Add to source-based top 5
		if len(metrics.TopLatencySrc) < 5 && !seen["src:"+info.SrcIP] {
			metrics.TopLatencySrc = append(metrics.TopLatencySrc, info)
			seen["src:"+info.SrcIP] = true
		}
		// Add to destination-based top 5
		if len(metrics.TopLatencyDst) < 5 && !seen["dst:"+info.DstIP] {
			metrics.TopLatencyDst = append(metrics.TopLatencyDst, info)
			seen["dst:"+info.DstIP] = true
		}

		if len(metrics.TopLatencySrc) >= 5 && len(metrics.TopLatencyDst) >= 5 {
			break
		}
	}

	return metrics
}

// resolveIP performs reverse DNS lookup and IP range detection
func (a *Aggregator) resolveIP(ip string) {
	// Skip localhost
	if ip == "127.0.0.1" || strings.HasPrefix(ip, "127.") {
		return
	}

	// Check cache first
	a.ipCacheMutex.RLock()
	if info, exists := a.ipCache[ip]; exists {
		// Refresh if older than 1 hour
		if time.Since(info.Resolved) < 1*time.Hour {
			a.ipCacheMutex.RUnlock()
			return
		}
	}
	a.ipCacheMutex.RUnlock()

	info := &types.IPInfo{
		IP:       ip,
		Resolved: time.Now(),
	}

	// Try reverse DNS lookup
	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Second)
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err == nil && len(names) > 0 {
		info.Hostname = strings.TrimSuffix(names[0], ".")
	}

	// Check known IP ranges for major companies
	info.Company = detectCompany(ip)

	// Cache the result
	a.ipCacheMutex.Lock()
	a.ipCache[ip] = info
	a.ipCacheMutex.Unlock()
}

// getResolvedName returns the best available name for an IP
func (a *Aggregator) getResolvedName(ip string) string {
	a.ipCacheMutex.RLock()
	defer a.ipCacheMutex.RUnlock()

	if info, exists := a.ipCache[ip]; exists {
		if info.Company != "" {
			return info.Company
		}
		if info.Hostname != "" {
			return info.Hostname
		}
	}

	return ip
}

// detectCompany checks if IP belongs to known company ranges
func detectCompany(ip string) string {
	// This is a simplified version. In production, you'd load comprehensive
	// IP range databases from AWS, GCP, Azure, etc.

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}

	// Google Cloud: 8.8.8.0/24, 8.8.4.0/24 (DNS), and many others
	if strings.HasPrefix(ip, "8.8.8.") || strings.HasPrefix(ip, "8.8.4.") {
		return "Google"
	}

	// AWS: Many ranges, this is just a sample
	// 52.0.0.0/8, 54.0.0.0/8, etc.
	if strings.HasPrefix(ip, "52.") || strings.HasPrefix(ip, "54.") {
		return "AWS"
	}

	// Cloudflare: 1.1.1.0/24, 1.0.0.0/24
	if strings.HasPrefix(ip, "1.1.1.") || strings.HasPrefix(ip, "1.0.0.") {
		return "Cloudflare"
	}

	// Add more ranges as needed
	return ""
}

// UpdateChannel returns the channel for topology updates
func (a *Aggregator) UpdateChannel() <-chan types.TopologyUpdate {
	return a.updateChan
}

// MetricsChannel returns the channel for metrics updates
func (a *Aggregator) MetricsChannel() <-chan types.MetricsUpdate {
	return a.metricsChan
}

// GetCurrentTopology returns the current complete topology
func (a *Aggregator) GetCurrentTopology() types.TopologyUpdate {
	a.flowsMutex.RLock()
	defer a.flowsMutex.RUnlock()

	flows := make([]types.Flow, 0, len(a.flows))
	for _, flow := range a.flows {
		flows = append(flows, *flow)
	}

	return types.TopologyUpdate{
		Timestamp: time.Now(),
		NewFlows:  flows,
		Summary:   a.calculateSummaryLocked(),
	}
}

// GetCurrentMetrics returns the current top-latency metrics snapshot
func (a *Aggregator) GetCurrentMetrics() types.MetricsUpdate {
	a.flowsMutex.RLock()
	defer a.flowsMutex.RUnlock()

	return a.calculateMetricsLocked()
}

// DumpTopology prints current topology (for debugging)
func (a *Aggregator) DumpTopology() {
	a.flowsMutex.RLock()
	defer a.flowsMutex.RUnlock()

	fmt.Println("\n=== Current Topology ===")
	fmt.Printf("Total flows: %d\n", len(a.flows))

	for _, flow := range a.flows {
		fmt.Printf("%s:%d -> %s:%d [%s] packets=%d bytes=%d latency=%.2fms\n",
			flow.SrcName, flow.Key.SrcPort,
			flow.DstName, flow.Key.DstPort,
			flow.ProtocolName,
			flow.Stats.Packets,
			flow.Stats.Bytes,
			flow.AvgLatencyMs,
		)
	}
	fmt.Println("========================\n")
}

// DumpJSON dumps current topology as JSON
func (a *Aggregator) DumpJSON() string {
	topology := a.GetCurrentTopology()
	data, _ := json.MarshalIndent(topology, "", "  ")
	return string(data)
}

// GetStore returns the storage instance
func (a *Aggregator) GetStore() *storage.Store {
	return a.store
}

// Close stops the aggregator
func (a *Aggregator) Close() {
	a.cancel()
	close(a.stopChan)
	close(a.updateChan)
	close(a.metricsChan)
	log.Println("Aggregator stopped")
}

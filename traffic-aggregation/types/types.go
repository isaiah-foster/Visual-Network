package types

import "time"

// Protocol constants - must match eBPF definitions
const (
	ProtoUnknown = 0
	ProtoHTTP    = 1
	ProtoHTTPS   = 2
	ProtoDNS     = 3
	ProtoSSH     = 4
)

// Connection states - must match eBPF definitions
const (
	StateNew         = 0
	StateEstablished = 1
	StateClosing     = 2
)

// FlowKey uniquely identifies a network flow
type FlowKey struct {
	SrcIP    uint32
	DstIP    uint32
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
	_        [3]byte // padding
}

// FlowStats contains statistics for a flow
type FlowStats struct {
	Packets      uint64
	Bytes        uint64
	LastSeen     uint64 // nanoseconds
	FirstSeen    uint64 // nanoseconds
	LatencySum   uint32 // microseconds
	LatencyCount uint32
	State        uint8
	L7Protocol   uint8
	_            [2]byte // padding
}

// ConnEvent is sent when a new connection is detected
type ConnEvent struct {
	SrcIP      uint32
	DstIP      uint32
	SrcPort    uint16
	DstPort    uint16
	Protocol   uint8
	L7Protocol uint8
	State      uint8
	_          byte // padding
	Timestamp  uint64
}

// Flow represents a complete network flow with computed fields
type Flow struct {
	Key          FlowKey
	Stats        FlowStats
	SrcIPStr     string
	DstIPStr     string
	SrcName      string // hostname or company name
	DstName      string
	ProtocolName string // "TCP", "UDP", "ICMP"
	L7ProtoName  string // "HTTP", "HTTPS", "DNS", etc.
	AvgLatencyMs float64
	Active       bool
	LastUpdate   time.Time
}

// TopologyUpdate represents changes to send to frontend
type TopologyUpdate struct {
	Timestamp time.Time      `json:"timestamp"`
	NewFlows  []Flow         `json:"new_flows,omitempty"`
	Updated   []Flow         `json:"updated,omitempty"`
	Removed   []FlowKey      `json:"removed,omitempty"`
	Summary   TopologySummary `json:"summary"`
}

// TopologySummary provides aggregate stats
type TopologySummary struct {
	TotalFlows   int     `json:"total_flows"`
	TotalPackets uint64  `json:"total_packets"`
	TotalBytes   uint64  `json:"total_bytes"`
	UniqueIPs    int     `json:"unique_ips"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

// LatencyInfo for top N highest latency flows
type LatencyInfo struct {
	SrcIP        string  `json:"src_ip"`
	DstIP        string  `json:"dst_ip"`
	SrcName      string  `json:"src_name"`
	DstName      string  `json:"dst_name"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	Protocol     string  `json:"protocol"`
}

// MetricsUpdate contains top latency info
type MetricsUpdate struct {
	Timestamp      time.Time     `json:"timestamp"`
	TopLatencySrc  []LatencyInfo `json:"top_latency_src"`  // top 5 by source
	TopLatencyDst  []LatencyInfo `json:"top_latency_dst"`  // top 5 by destination
}

// IPInfo stores hostname resolution results
type IPInfo struct {
	IP       string
	Hostname string
	Company  string // e.g., "AWS", "Google", etc.
	Resolved time.Time
}

// Helper functions
func (f *Flow) IsExpired(timeout time.Duration) bool {
	return time.Since(f.LastUpdate) > timeout
}

func ProtocolToString(proto uint8) string {
	switch proto {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 1:
		return "ICMP"
	default:
		return "UNKNOWN"
	}
}

func L7ProtocolToString(proto uint8) string {
	switch proto {
	case ProtoHTTP:
		return "HTTP"
	case ProtoHTTPS:
		return "HTTPS"
	case ProtoDNS:
		return "DNS"
	case ProtoSSH:
		return "SSH"
	default:
		return ""
	}
}

func StateToString(state uint8) string {
	switch state {
	case StateNew:
		return "NEW"
	case StateEstablished:
		return "ESTABLISHED"
	case StateClosing:
		return "CLOSING"
	default:
		return "UNKNOWN"
	}
}

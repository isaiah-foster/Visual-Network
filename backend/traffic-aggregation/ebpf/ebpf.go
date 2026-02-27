package ebpf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"visual-network/traffic-aggregation/types"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror -D__TARGET_ARCH_x86 -I/usr/include/x86_64-linux-gnu" tc_monitor ../../traffic-aggregation/ebpf/tc_monitor.c

// bpfFlowKey matches the C struct flow_key
type bpfFlowKey struct {
	SrcIp    uint32
	DstIp    uint32
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
	_        [3]byte
}

// bpfFlowStats matches the C struct flow_stats
type bpfFlowStats struct {
	Packets      uint64
	Bytes        uint64
	LastSeen     uint64
	FirstSeen    uint64
	LatencySum   uint32
	LatencyCount uint32
	State        uint8
	L7Protocol   uint8
	_            [2]byte
}

// Monitor manages the eBPF program and data collection
type Monitor struct {
	objs      *tc_monitorObjects
	link      link.Link
	ringbuf   *ringbuf.Reader
	iface     string
	eventChan chan types.ConnEvent
	stopChan  chan struct{}
}

// NewMonitor creates a new eBPF monitor
func NewMonitor(ifaceName string) (*Monitor, error) {
	// Load compiled eBPF objects
	objs := &tc_monitorObjects{}
	if err := loadTc_monitorObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("loading eBPF objects: %w", err)
	}

	// Get interface
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("getting interface %s: %w", ifaceName, err)
	}

	// Attach TC program to interface egress using netlink
	l, err := attachTCProgram(iface.Index, objs.TcMonitor)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("attaching TC program: %w", err)
	}

	// Open ring buffer for events
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		l.Close()
		objs.Close()
		return nil, fmt.Errorf("opening ring buffer: %w", err)
	}

	m := &Monitor{
		objs:      objs,
		link:      l,
		ringbuf:   rd,
		iface:     ifaceName,
		eventChan: make(chan types.ConnEvent, 100),
		stopChan:  make(chan struct{}),
	}

	// Start ring buffer reader
	go m.readEvents()

	log.Printf("eBPF monitor attached to interface %s", ifaceName)
	return m, nil
}

// attachTCProgram attaches a TC program to an interface using netlink
func attachTCProgram(ifaceIndex int, prog *ebpf.Program) (link.Link, error) {
	// Attach to TC ingress hook using TCX (TC eXtended) API
	opts := link.TCXOptions{
		Interface: ifaceIndex,
		Program:   prog,
		Attach:    ebpf.AttachTCXIngress,
	}

	l, err := link.AttachTCX(opts)
	if err == nil {
		return l, nil
	}

	// Fallback: Try attaching to egress if ingress fails
	opts.Attach = ebpf.AttachTCXEgress
	return link.AttachTCX(opts)
}

// readEvents reads from the ring buffer and sends to event channel
func (m *Monitor) readEvents() {
	for {
		select {
		case <-m.stopChan:
			return
		default:
			record, err := m.ringbuf.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					return
				}
				log.Printf("reading from ring buffer: %v", err)
				continue
			}

			// Parse event
			if len(record.RawSample) < 24 {
				continue
			}

			event := types.ConnEvent{
				SrcIP:      binary.LittleEndian.Uint32(record.RawSample[0:4]),
				DstIP:      binary.LittleEndian.Uint32(record.RawSample[4:8]),
				SrcPort:    binary.LittleEndian.Uint16(record.RawSample[8:10]),
				DstPort:    binary.LittleEndian.Uint16(record.RawSample[10:12]),
				Protocol:   record.RawSample[12],
				L7Protocol: record.RawSample[13],
				State:      record.RawSample[14],
				Timestamp:  binary.LittleEndian.Uint64(record.RawSample[16:24]),
			}

			select {
			case m.eventChan <- event:
			default:
				log.Printf("event channel full, dropping event")
			}
		}
	}
}

// EventChannel returns the channel for new connection events
func (m *Monitor) EventChannel() <-chan types.ConnEvent {
	return m.eventChan
}

// GetAllFlows reads all current flows from the flow map
func (m *Monitor) GetAllFlows() (map[types.FlowKey]types.FlowStats, error) {
	flows := make(map[types.FlowKey]types.FlowStats)

	var (
		key   bpfFlowKey
		value bpfFlowStats
	)

	iter := m.objs.FlowMap.Iterate()
	for iter.Next(&key, &value) {
		flowKey := types.FlowKey{
			SrcIP:    key.SrcIp,
			DstIP:    key.DstIp,
			SrcPort:  key.SrcPort,
			DstPort:  key.DstPort,
			Protocol: key.Protocol,
		}

		flowStats := types.FlowStats{
			Packets:      value.Packets,
			Bytes:        value.Bytes,
			LastSeen:     value.LastSeen,
			FirstSeen:    value.FirstSeen,
			LatencySum:   value.LatencySum,
			LatencyCount: value.LatencyCount,
			State:        value.State,
			L7Protocol:   value.L7Protocol,
		}

		flows[flowKey] = flowStats
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("iterating flow map: %w", err)
	}

	return flows, nil
}

// DeleteFlow removes a flow from the map
func (m *Monitor) DeleteFlow(key types.FlowKey) error {
	bpfKey := bpfFlowKey{
		SrcIp:    key.SrcIP,
		DstIp:    key.DstIP,
		SrcPort:  key.SrcPort,
		DstPort:  key.DstPort,
		Protocol: key.Protocol,
	}
	return m.objs.FlowMap.Delete(bpfKey)
}

// CleanExpiredFlows removes flows that haven't seen traffic in the timeout period
func (m *Monitor) CleanExpiredFlows(timeout time.Duration) (int, error) {
	now := time.Now()
	timeoutNs := uint64(timeout.Nanoseconds())
	nowNs := uint64(now.UnixNano())

	var (
		key   bpfFlowKey
		value bpfFlowStats
	)

	var toDelete []bpfFlowKey
	iter := m.objs.FlowMap.Iterate()
	for iter.Next(&key, &value) {
		if nowNs-value.LastSeen > timeoutNs {
			toDelete = append(toDelete, key)
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("iterating flow map: %w", err)
	}

	// Delete expired flows
	for _, k := range toDelete {
		if err := m.objs.FlowMap.Delete(k); err != nil {
			log.Printf("failed to delete flow: %v", err)
		}
	}

	return len(toDelete), nil
}

// GetMapStats returns statistics about the eBPF maps
func (m *Monitor) GetMapStats() (flowCount, handshakeCount int, err error) {
	var key, value interface{}

	// Count flows
	iter := m.objs.FlowMap.Iterate()
	for iter.Next(&key, &value) {
		flowCount++
	}
	if err := iter.Err(); err != nil {
		return 0, 0, err
	}

	// Count handshakes
	iter = m.objs.HandshakeMap.Iterate()
	for iter.Next(&key, &value) {
		handshakeCount++
	}
	if err := iter.Err(); err != nil {
		return 0, 0, err
	}

	return flowCount, handshakeCount, nil
}

// Close cleans up resources
func (m *Monitor) Close() error {
	close(m.stopChan)

	if m.ringbuf != nil {
		if err := m.ringbuf.Close(); err != nil {
			log.Printf("closing ring buffer: %v", err)
		}
	}

	if m.link != nil {
		if err := m.link.Close(); err != nil {
			log.Printf("closing TC link: %v", err)
		}
	}

	if m.objs != nil {
		m.objs.Close()
	}

	close(m.eventChan)
	log.Printf("eBPF monitor closed")
	return nil
}

// Helper functions
func Uint32ToIP(ip uint32) net.IP {
	result := make(net.IP, 4)
	binary.LittleEndian.PutUint32(result, ip)
	return result
}

func IPToUint32(ip net.IP) uint32 {
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(ip4)
}

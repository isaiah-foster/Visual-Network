//go:build ignore

#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define MAX_FLOWS 10000
#define FLOW_TIMEOUT_NS 10000000000ULL  // 10 seconds in nanoseconds

// L7 Protocol detection
#define PROTO_UNKNOWN 0
#define PROTO_HTTP 1
#define PROTO_HTTPS 2
#define PROTO_DNS 3
#define PROTO_SSH 4

// Connection states
#define STATE_NEW 0
#define STATE_ESTABLISHED 1
#define STATE_CLOSING 2

// Flow key - uniquely identifies a network flow
struct flow_key {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 protocol;  // IPPROTO_TCP, IPPROTO_UDP, etc.
    __u8 pad[3];
} __attribute__((packed));

// Flow statistics
struct flow_stats {
    __u64 packets;
    __u64 bytes;
    __u64 last_seen;
    __u64 first_seen;
    __u32 latency_sum;  // Sum of RTT measurements in microseconds
    __u32 latency_count; // Number of RTT samples
    __u8 state;
    __u8 l7_protocol;
    __u8 pad[2];
} __attribute__((packed));

// Connection event - sent to userspace when new connection detected
struct conn_event {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 protocol;
    __u8 l7_protocol;
    __u8 state;
    __u8 pad;
    __u64 timestamp;
} __attribute__((packed));

// TCP handshake tracking for RTT measurement
struct tcp_handshake {
    __u64 syn_time;
    __u32 syn_seq;
} __attribute__((packed));

// BPF Maps
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_FLOWS);
    __type(key, struct flow_key);
    __type(value, struct flow_stats);
} flow_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);  // 256KB ring buffer
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_FLOWS);
    __type(key, struct flow_key);
    __type(value, struct tcp_handshake);
} handshake_map SEC(".maps");

// Helper function to detect L7 protocol
static __always_inline __u8 detect_l7_protocol(__u16 port, __u8 *payload, __u32 payload_len) {
    // Port-based detection
    if (port == 80) return PROTO_HTTP;
    if (port == 443) return PROTO_HTTPS;
    if (port == 53) return PROTO_DNS;
    if (port == 22) return PROTO_SSH;
    
    // Payload-based detection (basic)
    if (payload_len >= 4) {
        // HTTP methods
        if (payload[0] == 'G' && payload[1] == 'E' && payload[2] == 'T' && payload[3] == ' ')
            return PROTO_HTTP;
        if (payload[0] == 'P' && payload[1] == 'O' && payload[2] == 'S' && payload[3] == 'T')
            return PROTO_HTTP;
        if (payload[0] == 'H' && payload[1] == 'T' && payload[2] == 'T' && payload[3] == 'P')
            return PROTO_HTTP;
    }
    
    return PROTO_UNKNOWN;
}

// Main TC classifier program
SEC("classifier")
int tc_monitor(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;
    
    // Only process IPv4 for now
    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return TC_ACT_OK;
    
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return TC_ACT_OK;
    
    // Get current timestamp
    __u64 now = bpf_ktime_get_ns();
    
    struct flow_key key = {};
    key.src_ip = ip->saddr;
    key.dst_ip = ip->daddr;
    key.protocol = ip->protocol;
    
    __u8 l7_proto = PROTO_UNKNOWN;
    __u32 packet_size = bpf_ntohs(ip->tot_len);
    void *l4_header = (void *)ip + (ip->ihl * 4);
    
    // Process TCP
    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = l4_header;
        if ((void *)(tcp + 1) > data_end)
            return TC_ACT_OK;
        
        key.src_port = bpf_ntohs(tcp->source);
        key.dst_port = bpf_ntohs(tcp->dest);
        
        // Get payload for L7 detection
        void *payload = (void *)tcp + (tcp->doff * 4);
        __u32 payload_len = 0;
        if (payload < data_end) {
            payload_len = data_end - payload;
            l7_proto = detect_l7_protocol(key.dst_port, payload, payload_len);
        }
        
        // Look up or create flow stats
        struct flow_stats *stats = bpf_map_lookup_elem(&flow_map, &key);
        if (!stats) {
            struct flow_stats new_stats = {};
            new_stats.packets = 1;
            new_stats.bytes = packet_size;
            new_stats.last_seen = now;
            new_stats.first_seen = now;
            new_stats.l7_protocol = l7_proto;
            
            // Determine initial state
            if (tcp->syn && !tcp->ack) {
                new_stats.state = STATE_NEW;
                
                // Track SYN for RTT measurement
                struct tcp_handshake hs = {};
                hs.syn_time = now;
                hs.syn_seq = bpf_ntohl(tcp->seq);
                bpf_map_update_elem(&handshake_map, &key, &hs, BPF_ANY);
                
            } else {
                new_stats.state = STATE_ESTABLISHED;
            }
            
            bpf_map_update_elem(&flow_map, &key, &new_stats, BPF_NOEXIST);
            
            // Send connection event to userspace
            struct conn_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
            if (event) {
                event->src_ip = key.src_ip;
                event->dst_ip = key.dst_ip;
                event->src_port = key.src_port;
                event->dst_port = key.dst_port;
                event->protocol = key.protocol;
                event->l7_protocol = l7_proto;
                event->state = new_stats.state;
                event->timestamp = now;
                bpf_ringbuf_submit(event, 0);
            }
        } else {
            // Update existing flow
            stats->packets++;
            stats->bytes += packet_size;
            stats->last_seen = now;
            
            // Update L7 protocol if we detected it now and didn't before
            if (l7_proto != PROTO_UNKNOWN && stats->l7_protocol == PROTO_UNKNOWN) {
                stats->l7_protocol = l7_proto;
            }
            
            // Track SYN-ACK for RTT calculation
            if (tcp->syn && tcp->ack) {
                // This is SYN-ACK response
                // Look for matching SYN
                struct flow_key reverse_key = {
                    .src_ip = key.dst_ip,
                    .dst_ip = key.src_ip,
                    .src_port = key.dst_port,
                    .dst_port = key.src_port,
                    .protocol = key.protocol
                };
                
                struct tcp_handshake *hs = bpf_map_lookup_elem(&handshake_map, &reverse_key);
                if (hs && hs->syn_time > 0) {
                    __u64 rtt_ns = now - hs->syn_time;
                    __u32 rtt_us = rtt_ns / 1000;
                    
                    // Update latency stats
                    stats->latency_sum += rtt_us;
                    stats->latency_count++;
                    
                    // Update state
                    stats->state = STATE_ESTABLISHED;
                    
                    // Clean up handshake tracking
                    bpf_map_delete_elem(&handshake_map, &reverse_key);
                }
            }
            
            // Track FIN for connection closing
            if (tcp->fin) {
                stats->state = STATE_CLOSING;
            }
        }
    }
    // Process UDP
    else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = l4_header;
        if ((void *)(udp + 1) > data_end)
            return TC_ACT_OK;
        
        key.src_port = bpf_ntohs(udp->source);
        key.dst_port = bpf_ntohs(udp->dest);
        
        // Get payload for L7 detection
        void *payload = (void *)(udp + 1);
        __u32 payload_len = 0;
        if (payload < data_end) {
            payload_len = data_end - payload;
            l7_proto = detect_l7_protocol(key.dst_port, payload, payload_len);
        }
        
        struct flow_stats *stats = bpf_map_lookup_elem(&flow_map, &key);
        if (!stats) {
            struct flow_stats new_stats = {};
            new_stats.packets = 1;
            new_stats.bytes = packet_size;
            new_stats.last_seen = now;
            new_stats.first_seen = now;
            new_stats.state = STATE_ESTABLISHED;  // UDP is connectionless
            new_stats.l7_protocol = l7_proto;
            
            bpf_map_update_elem(&flow_map, &key, &new_stats, BPF_NOEXIST);
            
            // Send event
            struct conn_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
            if (event) {
                event->src_ip = key.src_ip;
                event->dst_ip = key.dst_ip;
                event->src_port = key.src_port;
                event->dst_port = key.dst_port;
                event->protocol = key.protocol;
                event->l7_protocol = l7_proto;
                event->state = STATE_ESTABLISHED;
                event->timestamp = now;
                bpf_ringbuf_submit(event, 0);
            }
        } else {
            stats->packets++;
            stats->bytes += packet_size;
            stats->last_seen = now;
            
            if (l7_proto != PROTO_UNKNOWN && stats->l7_protocol == PROTO_UNKNOWN) {
                stats->l7_protocol = l7_proto;
            }
        }
    }
    // Process ICMP
    else if (ip->protocol == IPPROTO_ICMP) {
        // For ICMP, we use type/code as "ports"
        __u8 *icmp_data = l4_header;
        if (icmp_data + 2 > (__u8 *)data_end)
            return TC_ACT_OK;
        
        key.src_port = icmp_data[0];  // ICMP type
        key.dst_port = icmp_data[1];  // ICMP code
        
        struct flow_stats *stats = bpf_map_lookup_elem(&flow_map, &key);
        if (!stats) {
            struct flow_stats new_stats = {};
            new_stats.packets = 1;
            new_stats.bytes = packet_size;
            new_stats.last_seen = now;
            new_stats.first_seen = now;
            new_stats.state = STATE_ESTABLISHED;
            new_stats.l7_protocol = PROTO_UNKNOWN;
            
            bpf_map_update_elem(&flow_map, &key, &new_stats, BPF_NOEXIST);
            
            struct conn_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
            if (event) {
                event->src_ip = key.src_ip;
                event->dst_ip = key.dst_ip;
                event->src_port = key.src_port;
                event->dst_port = key.dst_port;
                event->protocol = key.protocol;
                event->l7_protocol = PROTO_UNKNOWN;
                event->state = STATE_ESTABLISHED;
                event->timestamp = now;
                bpf_ringbuf_submit(event, 0);
            }
        } else {
            stats->packets++;
            stats->bytes += packet_size;
            stats->last_seen = now;
        }
    }
    
    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";

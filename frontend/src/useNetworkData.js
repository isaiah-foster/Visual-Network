import { useEffect, useMemo, useRef, useState } from 'react';

const API_BASE = process.env.REACT_APP_BACKEND_API_BASE || '/api/backend';
const EMPTY_SNAPSHOT_GRACE_MS = 15000;

function estimateFlowLatencyMs(flow) {
  const explicit = Number(flow?.avgLatencyMs || 0);
  if (Number.isFinite(explicit) && explicit > 0) {
    return explicit;
  }

  const firstSeen = Number(flow?.firstSeen || 0);
  const lastSeen = Number(flow?.lastSeen || 0);
  const packets = Number(flow?.packets || 0);

  if (lastSeen > firstSeen && packets > 1) {
    const deltaNs = lastSeen - firstSeen;
    const estimateMs = deltaNs / (packets - 1) / 1_000_000;
    if (Number.isFinite(estimateMs) && estimateMs > 0) {
      return estimateMs;
    }
  }

  return 0;
}

function buildEstimatedLatency(flows, direction) {
  const grouped = new Map();

  flows.forEach((flow) => {
    // Avoid ICMP-based 1s ping interval looking like RTT latency
    if ((flow?.protocolName || '').toUpperCase() !== 'TCP') {
      return;
    }

    const latencyMs = estimateFlowLatencyMs(flow);
    if (!Number.isFinite(latencyMs) || latencyMs <= 0) {
      return;
    }

    const key = direction === 'src'
      ? (flow?.srcLabel || flow?.srcName || 'Unknown')
      : (flow?.dstLabel || flow?.dstName || 'Unknown');

    const current = grouped.get(key) || { sum: 0, count: 0 };
    grouped.set(key, {
      sum: current.sum + latencyMs,
      count: current.count + 1,
    });
  });

  return Array.from(grouped.entries())
    .map(([name, v]) => ({ name, avgLatencyMs: v.sum / v.count }))
    .sort((a, b) => b.avgLatencyMs - a.avgLatencyMs)
    .slice(0, 5);
}

function normalizeFlow(flow) {
  const stats = flow?.Stats || flow?.stats || {};
  const srcIP = flow?.SrcIPStr ?? flow?.src_ip_str ?? '';
  const dstIP = flow?.DstIPStr ?? flow?.dst_ip_str ?? '';
  const srcName = flow?.SrcName ?? flow?.src_name ?? srcIP ?? 'Unknown';
  const dstName = flow?.DstName ?? flow?.dst_name ?? dstIP ?? 'Unknown';
  const srcLabel = srcName !== srcIP && srcIP ? `${srcName} (${srcIP})` : srcName;
  const dstLabel = dstName !== dstIP && dstIP ? `${dstName} (${dstIP})` : dstName;

  return {
    srcName,
    dstName,
    srcIP,
    dstIP,
    srcLabel,
    dstLabel,
    protocolName: flow?.ProtocolName ?? flow?.protocol_name ?? 'UNKNOWN',
    l7ProtoName: flow?.L7ProtoName ?? flow?.l7_proto_name ?? '',
    bytes: Number(stats?.Bytes ?? stats?.bytes ?? 0),
    packets: Number(stats?.Packets ?? stats?.packets ?? 0),
    firstSeen: Number(stats?.FirstSeen ?? stats?.first_seen ?? 0),
    lastSeen: Number(stats?.LastSeen ?? stats?.last_seen ?? 0),
    avgLatencyMs: Number(flow?.AvgLatencyMs ?? flow?.avg_latency_ms ?? 0),
  };
}

function normalizeLatency(item, direction) {
  const sourceName = item?.SrcName ?? item?.src_name ?? item?.SrcIP ?? item?.src_ip ?? 'Unknown';
  const destinationName = item?.DstName ?? item?.dst_name ?? item?.DstIP ?? item?.dst_ip ?? 'Unknown';

  return {
    name: direction === 'src' ? sourceName : destinationName,
    avgLatencyMs: Number(item?.AvgLatencyMs ?? item?.avg_latency_ms ?? 0)
  };
}

async function fetchJson(path) {
  const response = await fetch(`${API_BASE}${path}`);
  const contentType = response.headers.get('content-type') || '';

  if (!response.ok) {
    throw new Error(`Request failed (${response.status}) for ${path}`);
  }

  if (!contentType.toLowerCase().includes('application/json')) {
    const body = await response.text();
    const preview = body.slice(0, 120).replace(/\s+/g, ' ');
    throw new Error(
      `Expected JSON from ${path}, got ${contentType || 'unknown content type'} (${preview || 'empty body'})`
    );
  }

  return response.json();
}

export function useNetworkData(pollMs = 3000) {
  const [topology, setTopology] = useState(null);
  const [metrics, setMetrics] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const lastNonEmptyTopologyRef = useRef(null);
  const lastNonEmptyAtRef = useRef(0);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        const [nextTopologyRaw, nextMetricsRaw] = await Promise.all([
          fetchJson('/topology'),
          fetchJson('/metrics')
        ]);

        const rawFlows = nextTopologyRaw?.NewFlows || nextTopologyRaw?.new_flows || [];
        const normalizedTopology = {
          ...nextTopologyRaw,
          flows: rawFlows.map(normalizeFlow)
        };

        const now = Date.now();
        let topologyToSet = normalizedTopology;
        if (normalizedTopology.flows.length > 0) {
          lastNonEmptyTopologyRef.current = normalizedTopology;
          lastNonEmptyAtRef.current = now;
        } else if (
          lastNonEmptyTopologyRef.current &&
          now-lastNonEmptyAtRef.current <= EMPTY_SNAPSHOT_GRACE_MS
        ) {
          topologyToSet = {
            ...normalizedTopology,
            flows: lastNonEmptyTopologyRef.current.flows,
            summary: lastNonEmptyTopologyRef.current.summary || normalizedTopology.summary,
            Summary: lastNonEmptyTopologyRef.current.Summary || normalizedTopology.Summary,
          };
        }

        const srcLatencyRaw = nextMetricsRaw?.TopLatencySrc || nextMetricsRaw?.top_latency_src || [];
        const dstLatencyRaw = nextMetricsRaw?.TopLatencyDst || nextMetricsRaw?.top_latency_dst || [];
        const normalizedMetrics = {
          ...nextMetricsRaw,
          topLatencySrc: srcLatencyRaw.map((item) => normalizeLatency(item, 'src')),
          topLatencyDst: dstLatencyRaw.map((item) => normalizeLatency(item, 'dst'))
        };

        if (!cancelled) {
          setTopology(topologyToSet);
          setMetrics(normalizedMetrics);
          setError('');
          setLoading(false);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load backend network data');
          setLoading(false);
        }
      }
    };

    load();
    const timer = setInterval(load, pollMs);

    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [pollMs]);

  const flows = useMemo(() => topology?.flows || [], [topology]);
  const latencySrc = useMemo(() => {
    const measured = metrics?.topLatencySrc || [];
    if (measured.length > 0) {
      return measured;
    }
    return buildEstimatedLatency(flows, 'src');
  }, [metrics, flows]);

  const latencyDst = useMemo(() => {
    const measured = metrics?.topLatencyDst || [];
    if (measured.length > 0) {
      return measured;
    }
    return buildEstimatedLatency(flows, 'dst');
  }, [metrics, flows]);

  return { topology, metrics, flows, latencySrc, latencyDst, loading, error };
}

# Sprint 3 Report (4/5/26 - 5/2/2026)

## What's New (User Facing)

### Admin Features
* Feature 1: The traffic aggregation backend can now be secured with an API key, preventing unauthorized access when the collector is exposed on a network interface other than localhost.
* Feature 2: The backend and frontend can now be deployed on separate physical machines. The collector runs on the monitored host (requiring only root for eBPF), while the reverse proxy, auth service, and frontend run on a separate management host with no elevated privileges required.

### General User Features
* Feature 1: Network latency data is now significantly more accurate. Latency is computed exclusively from real TCP handshake RTT measurements rather than relying on a fallback that incorrectly reported inter-packet intervals as latency.
* Feature 2: Latency metrics now include p50, p95, and p99 percentile breakdowns per flow in addition to the average, giving a better picture of tail latency rather than just the mean.
* Feature 3: Network data now arrives in the browser in real time via WebSocket push from the reverse proxy, replacing the previous 3-second HTTP polling cycle. The browser falls back to polling automatically if the WebSocket is unavailable.
* Feature 4: The frontend UX and visual design have been significantly improved. The layout has been restructured to eliminate the need to scroll to see network protocol charts on laptop-sized screens, the background and color palette have been updated from the placeholder gray-and-white scheme, and the home page content has been reorganized into clearly separated containers that make the application easier to understand at a glance. (Issue #6, carried over from sprint 2.)


## Work Summary (Developer Facing)

### Issue #6 — Frontend UX and Visual Design (Resolved, carried from Sprint 2)
The frontend received a visual overhaul that was code-complete at the end of sprint 2 but had not yet been merged. This sprint it was merged into main. The layout was reworked so that all key charts and the network protocol breakdown are visible without scrolling on standard laptop displays. The color scheme was replaced with a consistent, production-ready palette. The home page was reorganized from a single dense text block into individual informational containers, improving clarity for new users.

### Issue #17 — Improve Latency Data Flows from Backend
The previous latency implementation had several correctness issues that this sprint addresses:

1. **Overflow fix**: `latency_sum` in the eBPF `flow_stats` struct was a 32-bit unsigned integer, which silently wraps around after accumulating roughly 4.3 seconds worth of RTT measurements. It has been widened to 64 bits both in the C struct and in all corresponding Go types.

2. **Removed misleading fallback**: When a flow had no measured RTT data (e.g., UDP), the aggregator previously substituted `(last_seen - first_seen) / packets` as an estimated latency. This value is actually the average inter-packet interval — not latency — and produced wildly incorrect numbers. That code path has been removed. Flows without measured RTT data now correctly report zero latency.

3. **Percentile tracking via ring buffer**: A second eBPF ring buffer (`rtt_events`) was added alongside the existing connection-event buffer. Each time a TCP SYN-ACK RTT is measured in the kernel, an individual sample is emitted. The Go aggregator maintains a per-flow sliding window of up to 100 samples and computes p50, p95, p99, and jitter (standard deviation) on each aggregation cycle. These values are included in the `/metrics` API response and surfaced in the frontend latency charts.

### Issue #13 — Support Separation of Backend and Frontend Hosts
The reverse proxy already supported a configurable `-backend` URL flag, but several gaps made true multi-host deployment impractical:

1. **API key authentication on the aggregator**: The collector now reads `COLLECTOR_API_KEY` from the environment. When set, all endpoints require an `Authorization: Bearer <key>` header. When unset, the behavior is unchanged for local development. CORS headers are also added, configurable via `ALLOWED_ORIGINS`.

2. **WebSocket pipeline connected to real data**: The `MetricsHub` and `/api/metrics/ws` endpoint existed in the reverse proxy since sprint 2 but were never wired to actual aggregator data. A `startBackendPoller` goroutine was added to the reverse proxy that polls `/topology` and `/metrics` from the aggregator every second, forwarding the API key, and broadcasts the combined payload to all connected WebSocket clients. This enables the reverse proxy to act as the sole public-facing entry point even when the aggregator is on a different host.

3. **Frontend WebSocket-first consumption**: `useNetworkData.js` now attempts a WebSocket connection to `/api/metrics/ws` on mount. Incoming messages are parsed and applied immediately. HTTP polling continues as a fallback and activates automatically when the WebSocket is unavailable or disconnected.

4. **Makefile deployment targets**: `run-collector-remote` and `run-proxy-remote` targets document and automate the split-host deployment pattern.


## Unfinished Work
* Admin user management features (editing accounts, removing users, forced password reset on first login) mentioned in sprint 2 retrospective were not addressed this sprint.
* Active latency probing (ping from frontend to test latency to arbitrary servers) was planned in sprint 2 but not implemented.


## Completed Issues/User Stories
Here are links to the issues completed in this sprint:

* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/6
* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/17
* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/13


## Incomplete Issues/User Stories
Here are links to issues we worked on but did not complete in this sprint:

* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/5 (Active latency probing / ping from frontend — still planned)


## Code Files for Review (top 5-6 files that are highlight / best files)
Please review the following code files, which were actively developed during this sprint, for quality:

* [tc_monitor.c](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/ebpf/tc_monitor.c) — eBPF kernel program: u64 latency_sum fix, new `rtt_events` ring buffer and `rtt_event` struct
* [ebpf.go](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/ebpf/ebpf.go) — Go eBPF wrapper: RTT ring buffer reader, `RTTEventChannel` exposure
* [aggregator.go](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/aggregator/aggregator.go) — Per-flow RTT sample window, percentile/jitter computation, fallback removal
* [main.go (traffic-aggregation)](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/main.go) — API key middleware and CORS middleware
* [main.go (rev-proxy)](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/rev-proxy/main.go) — `startBackendPoller` goroutine connecting the MetricsHub to live aggregator data
* [useNetworkData.js](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/frontend/src/useNetworkData.js) — WebSocket-first data consumption with HTTP polling fallback


## Retrospective Summary

Here's what went well:
* The latency accuracy improvements were clean to implement because the eBPF ring buffer pattern was already established from sprint 2 — adding a second buffer for RTT samples required minimal scaffolding.
* The WebSocket pipeline came together quickly once the existing `MetricsHub` infrastructure in the reverse proxy was properly wired to a data source. The foundation from sprint 2 paid off.
* Split-host deployment required no architectural changes to the aggregator itself, only the addition of middleware — a sign that the original service boundary was drawn in the right place.

Here's what we'd like to improve:
* The sprint 3 report was partially empty at the start of the sprint, suggesting team members should fill it in incrementally rather than at the end.
* Admin user management features have been deferred across three sprints and should be prioritized or formally dropped.
* UDP and ICMP latency remain unmeasured; if latency for non-TCP flows matters to users, a follow-up issue should track it explicitly.

Here are changes we plan to implement going forward:
* Admin account management (edit user info, remove users, forced password change on first login).
* Active latency probing from the frontend (ICMP echo / ping to user-specified hosts).
* Evaluate whether p95/p99 latency data should be surfaced more prominently in the frontend charts.
* Harden the split-host deployment with TLS between the aggregator and the reverse proxy.

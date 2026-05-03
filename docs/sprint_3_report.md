# Sprint 3 Report (4/5/26 - 5/2/2026)

## What's New (User Facing)

### Admin Features
* Feature 1: API key authentication on collector to prevent unauthorized access
* Feature 2: Backend and frontend can now be deployed on separate physical machines

### General User Features
* Feature 1: Network latency data is now accurate (computed from TCP RTT only, not inter-packet intervals)
* Feature 2: Latency metrics now include p50, p95, and p99 percentiles in addition to average
* Feature 3: Real-time data delivery via WebSocket push (with HTTP polling fallback)
* Feature 4: Frontend UX and design overhaul with improved layout and color scheme


## Work Summary (Developer Facing)
We completed three major issues: frontend UX redesign (issue #6), improved latency accuracy (issue #17), and multi-host deployment support (issue #13). Latency accuracy improvements included fixing a 32-bit overflow in the eBPF struct, removing a misleading fallback calculation, and adding RTT percentile tracking via a second ring buffer. Multi-host deployment required API key middleware on the aggregator, wiring the MetricsHub to live data with a backend poller, and WebSocket-first consumption on the frontend with polling fallback.


## Unfinished Work
* Admin user management features (editing accounts, removing users, forced password reset)
* Active latency probing (ping from frontend to arbitrary servers)


## Completed Issues/User Stories
* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/6
* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/17
* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/13


## Incomplete Issues/User Stories
* https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/5 (Active latency probing)


## Code Files for Review
* [tc_monitor.c](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/ebpf/tc_monitor.c)
* [ebpf.go](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/ebpf/ebpf.go)
* [aggregator.go](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/aggregator/aggregator.go)
* [main.go (traffic-aggregation)](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/main.go)
* [main.go (rev-proxy)](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/rev-proxy/main.go)
* [useNetworkData.js](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/frontend/src/useNetworkData.js)


## Retrospective Summary

Here's what went well:
* eBPF ring buffer pattern from sprint 2 made adding RTT tracking straightforward
* WebSocket pipeline connected quickly once existing infrastructure was wired to data
* Multi-host deployment required no architectural changes, only middleware

Here's what we'd like to improve:
* Fill in sprint report incrementally, not at deadline
* Admin features deferred three sprints — prioritize or drop
* UDP/ICMP latency remain unmeasured

Here are changes we plan to implement in the next sprint:
* Admin account management (edit user, remove users, forced password change)
* Active latency probing from frontend
* Evaluate surfacing p95/p99 latency more prominently
* Harden multi-host deployment with TLS

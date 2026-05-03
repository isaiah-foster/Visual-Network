# Visual Network

## Project summary

### One-sentence description of the project

Visual Network is a full-stack application that captures live network traffic on Linux via eBPF and visualizes inbound, outbound, and mixed traffic patterns through an authenticated React dashboard.

### Additional information about the project

Visual Network is made up of four services:

- **Frontend** — a React SPA with login-gated routes that visualizes packet volume, protocol distribution, and latency using pie and bar charts.
- **Backend collector** — a Go program that attaches eBPF/TC programs to a network interface to capture and aggregate traffic in real time.
- **Auth service** — a Go service backed by MySQL that handles login, sessions, and role-based access (regular users vs. admins who can create new accounts).
- **Reverse proxy** — a Go server that serves the built frontend and proxies `/api/auth/*` and `/api/backend/*` to the auth service and collector, optionally over HTTPS with WebSocket support for live metrics.

The app provides Home, In, Out, and Mixed views. Admin users additionally get a "Add User" page for provisioning new accounts.

## Installation

### Prerequisites

- Linux kernel >= 6.6 (required for the eBPF/TC traffic collector)
- Go 1.25 or newer
- Node.js 20.x or newer
- npm 10.x or newer
- `clang`, `llvm` (for `llvm-strip`), `libbpf-dev`, and matching `linux-headers-$(uname -r)` — `make setup` will offer to install these for you via `backend/setup.sh`
- MySQL server (for the auth service — see [Auth service / database setup](#auth-service--database-setup))
- Git

### Installation steps

```bash
git clone https://github.com/WSU-CPTS322-SP26/Visual-Network.git
cd Visual-Network
make setup
make build
```

`make setup` installs frontend (npm), backend (Go modules + eBPF toolchain via `backend/setup.sh`), and reverse proxy dependencies. `make build` generates the eBPF bindings and builds the collector, auth service, reverse proxy, and frontend production bundle.

### Auth service / database setup

The auth service requires a MySQL database. From `backend/auth_service/`:

```bash
make mysql-ubuntu                                             # install/configure MySQL (Ubuntu)
make mysql-create-db                                          # create DB + schema from db/db_init.sql
make mysql-create-user DB_USER=<user> DB_PASS=<pass> DB_NAME=<db>
make .env-file DB_USER=<user> DB_PASS=<pass> DB_NAME=<db>      # writes .env with MYSQL_DSN
make seed-admin                                                # creates the first admin account
```

## Functionality

Run the full stack in four separate terminals (also printed by `make run-full`):

```bash
# Terminal 1 — Auth service (needs .env from setup above)
make run-auth_service

# Terminal 2 — Backend collector (needs a real network interface, requires sudo for eBPF)
ip link show                 # find your interface, e.g. eth0/wlan0
make run-backend IFACE=eth0

# Terminal 3 — Reverse proxy
make run-proxy

# Terminal 4 — Frontend dev server
make run-frontend
```

Then open the app:

1. Navigate to `http://localhost:3000` (dev server) or `http://localhost:8080` (via the reverse proxy, once the frontend is built).
2. Log in at `/login` — routes are protected and redirect unauthenticated users here.
3. Use the top navigation:
   - **Home**: Overview of the project and tracked metrics.
   - **In Page**: Inbound traffic focus (Volume In, Protocol, Latency).
   - **Out Page**: Outbound traffic focus (Volume Out, Protocol, Latency).
   - **Mixed Page**: Combined in/out perspective with protocol and latency charts.
   - **Add User** (admin only): Create new user accounts.
4. Hover chart segments/bars to inspect values.

For HTTPS, run `make setup-reverse-proxy` to generate self-signed certs, then access via `https://localhost:8443`.

### Split-host deployment

The collector and reverse proxy can run on separate hosts (collector needs root/eBPF privileges; the management host doesn't):

```bash
# Host A (network monitor)
make run-collector-remote IFACE=eth0 COLLECTOR_API_KEY=secret

# Host B (management host)
make run-proxy-remote COLLECTOR_HOST=<host-a-ip> COLLECTOR_API_KEY=secret
```

### Environment variables

- `MYSQL_DSN` — MySQL connection string used by the auth service and (optionally) the reverse proxy for session verification.
- `REACT_APP_AUTH_URL` — base URL the frontend uses for auth requests (login, logout, session check, admin user creation).
- `REACT_APP_BACKEND_API_BASE` — base URL for backend metrics API calls (defaults to `/api/backend`).

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

## Additional Documentation

Documentation is in `docs/`, including sprint reports (1-3) and demo materials for each sprint.

## License
[MIT License](https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/LICENSE.txt)

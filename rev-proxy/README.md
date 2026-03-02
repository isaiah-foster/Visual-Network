# Visual Network Reverse Proxy

A reverse proxy server that:
- Routes HTTPS traffic to frontend and backend services
- Handles authentication and session verification
- Provides WebSocket endpoints for real-time metrics streaming
- Proxies API requests to auth and backend services

## Setup

```bash
make setup
make build
```

## Running

### Without HTTPS (development):
```bash
make run-no-auth
```

### With HTTPS (production):
First, generate self-signed certificates (dev) or use real ones:
```bash
make generate-certs
```

Then run with MySQL for session verification:
```bash
MYSQL_DSN="user:pass@tcp(localhost:3306)/dbname" make run
```

## Architecture

### Routes

- `/` - Frontend (serves SPA from build directory)
- `/api/auth/*` - Proxied to auth service (localhost:8080)
- `/api/backend/*` - Proxied to backend services (localhost:9090)
- `/api/metrics/ws` - WebSocket for real-time metrics (requires auth if MySQL configured)
- `/health` - Health check endpoint

### Ports

- `8080` - HTTP (always running)
- `8443` - HTTPS (optional, if TLS cert/key provided)

### WebSocket Connection Example

```javascript
// Connect to metrics WebSocket
const ws = new WebSocket('wss://localhost:8443/api/metrics/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Metrics update:', message);
};

ws.onclose = () => {
  console.log('Disconnected from metrics');
};
```

## Environment Variables

- `MYSQL_DSN` - MySQL connection string (optional, for auth verification)

## Certificate Installation (TLS)

For production:
1. Obtain valid SSL/TLS certificates
2. Place cert.pem and key.pem in the certs/ directory
3. Run with `-cert certs/cert.pem -key certs/key.pem`

For development (self-signed):
```bash
make generate-certs
```

## Integration with Frontend

Add to frontend's `.env.production`:
```
REACT_APP_API_URL=https://localhost:8443/api
REACT_APP_METRICS_WS=wss://localhost:8443/api/metrics/ws
```

Or update API client to use:
- Auth endpoints: `https://localhost:8443/api/auth/*`
- Metrics WebSocket: `wss://localhost:8443/api/metrics/ws`

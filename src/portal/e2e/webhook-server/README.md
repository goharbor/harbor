# Webhook Server for E2E Tests

This directory contains the Docker Compose setup for a webhook receiver server used by Harbor E2E tests.

## Overview

The webhook server uses [webhook.site](https://github.com/webhooksite/webhook.site) - an open-source tool that captures and displays HTTP requests. It's used to verify that Harbor correctly sends webhook notifications for various events.

## Quick Start

### 1. Start the Webhook Server

```bash
cd src/portal/e2e/webhook-server
docker compose up -d
```

### 2. Verify it's Running

Open http://localhost:8084 in your browser. You should see the webhook.site interface with a unique webhook URL.

### 3. Run Webhook Tests

```bash
cd src/portal
WEBHOOK_ENDPOINT_UI=localhost:8084 IP=<harbor-ip> npx playwright test webhook.spec.ts
```

### 4. Stop the Server (when done)

```bash
docker compose down
```

## Services

| Service | Description | Port |
|---------|-------------|------|
| `webhook` | Main webhook receiver with web UI | 8084 |
| `redis` | Cache and queue backend | - |
| `laravel-echo-server` | Real-time updates via WebSocket | 6001 |

## Environment Variables

Set these when running Playwright tests:

| Variable | Description | Example |
|----------|-------------|---------|
| `WEBHOOK_ENDPOINT_UI` | Webhook server address | `localhost:8084` |
| `IP` | Harbor server IP/hostname | `192.168.1.100` |
| `HARBOR_ADMIN` | Harbor admin username | `admin` |
| `HARBOR_PASSWORD` | Harbor admin password | `Harbor12345` |

## How it Works

1. The test opens the webhook.site UI and retrieves the unique webhook URL
2. Harbor is configured to send webhooks to this URL
3. When Harbor triggers events (tag retention, replication, etc.), it sends HTTP POST requests
4. The webhook server captures and displays these requests
5. The test verifies the payload contains expected data

## Keeping the Server Running

For development, you can keep the webhook server running continuously:

```bash
# Start in background (survives terminal close)
docker compose up -d

# View logs
docker compose logs -f webhook

# Check status
docker compose ps
```

The server uses minimal resources and can stay running indefinitely.

## Troubleshooting

### Server not starting
```bash
# Check logs
docker compose logs webhook

# Restart services
docker compose restart
```

### Port 8084 already in use
```bash
# Find process using port
lsof -i :8084

# Or change the port in docker-compose.yml
ports:
  - "9084:80"  # Use port 9084 instead
```

### Webhook not receiving requests
1. Ensure Harbor can reach the webhook server (network connectivity)
2. Check Harbor webhook configuration points to correct URL
3. Verify the webhook is enabled in Harbor project settings

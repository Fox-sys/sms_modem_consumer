# sender-modem

Polls a GSM modem via HiLink API and forwards SMS to a remote HTTP API. Messages are deleted from the modem after reading.

**Tested with:** Huawei E3372-325.

## Stack

- Go 1.21+
- [cleanenv](https://github.com/ilyakaznacheev/cleanenv) — config from environment variables
- Build: multi-stage Docker (Alpine)

## Configuration

Environment variables only (prefix `SMS_`):

| Variable | Default | Description |
|----------|---------|-------------|
| `SMS_POLL_INTERVAL_SECONDS` | `60` | Poll interval in seconds |
| `SMS_MODEM_BASE_URL` | `http://192.168.8.1` | Modem base URL |
| `SMS_MODEM_USERNAME` | `admin` | HiLink username |
| `SMS_MODEM_PASSWORD` | `admin` | HiLink password |
| `SMS_API_BASE_URL` | — | Remote API base URL (required) |
| `SMS_API_KEY` | — | Bearer token for API (optional) |
| `SMS_LOG_LEVEL` | `info` | Log level: `info` or `debug`. With `debug`, each forwarded message is also printed to stdout (phone, content, date). |

API payload: `POST {SMS_API_BASE_URL}/api/sms` with a JSON array of objects `{ "index", "phone", "content", "date", "smstat", "sms_type" }`.

## Run

**Local (Go installed):**

```bash
go run ./src/cmd/sms-consumer
```

Set env vars before running (e.g. `export SMS_API_BASE_URL=...`) or use a `.env` file with a tool that injects them.

**Docker:**

Build (context is repo root; dependencies are installed before copying code so that layer is cached):

```bash
docker build -f deployment/Dockerfile -t sms_modem_consumer .
```

Run (modem on host network, e.g. `192.168.8.1`):

```bash
docker run --rm --network host \
  -e SMS_API_BASE_URL=https://your-api.example.com \
  -e SMS_API_KEY=your-token \
  sms_modem_consumer
```

Optional: `-e SMS_LOG_LEVEL=debug` to print each forwarded message (phone, content, date) to stdout.

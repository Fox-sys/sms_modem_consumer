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

API payload: `POST {SMS_API_BASE_URL}/api/sms` with a JSON array of objects `{ "index", "phone", "content", "date", "smstat", "sms_type" }`.

## Run

**Local (Go installed):**

```bash
cd src
go run ./cmd/sms-consumer
```

**Docker:**

Build (context is the `src` directory):

```bash
docker build -f deployment/Dockerfile -t sender-modem src
```

Run:

```bash
docker run --rm \
  -e SMS_API_BASE_URL=https://your-api.example.com \
  -e SMS_API_KEY=your-token \
  sender-modem
```

Use `--network host` or port mapping when the modem is on the local network.

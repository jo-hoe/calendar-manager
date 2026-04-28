# calendar-manager

[![Lint](https://github.com/jo-hoe/calendar-manager/actions/workflows/lint.yml/badge.svg)](https://github.com/jo-hoe/calendar-manager/actions/workflows/lint.yml)
[![Test](https://github.com/jo-hoe/calendar-manager/actions/workflows/test.yml/badge.svg)](https://github.com/jo-hoe/calendar-manager/actions/workflows/test.yml)

A Go HTTP server that manages calendar events via a REST API. External clients POST events to this server; it deduplicates against existing entries (via iCal URL) and creates or updates events on the configured calendar provider.

Currently supports Google Calendar. The provider interface allows adding more backends in the future.

## API

| Method   | Path              | Description                    |
|----------|-------------------|--------------------------------|
| `GET`    | `/probe`          | Health check                   |
| `GET`    | `/api/events`     | List events from iCal feed     |
| `POST`   | `/api/events`     | Create or update (upsert) event|
| `DELETE`  | `/api/events/{id}`| Delete an event                |

### POST /api/events

```json
{
  "title": "Union Berlin Frauen vs TSG Hoffenheim",
  "startTime": "2026-05-10T16:00:00+02:00",
  "endTime": "2026-05-10T18:00:00+02:00",
  "location": "STADION An der Alten Försterei, Berlin",
  "description": "Frauen-Bundesliga | 25. Spieltag"
}
```

**Deduplication:** If an event with the same title and start time already exists in the calendar, it is updated instead of duplicated.

## Configuration

```yaml
port: 8080
logLevel: "info"
calendar:
  provider: "google"
  google:
    credentialsFile: "/secrets/google-credentials.json"
    calendarID: "your-calendar-id@group.calendar.google.com"
  icalURL: "https://calendar.google.com/calendar/ical/your-calendar-id/basic.ics"
```

The server reads configuration from a YAML file. The path is determined by:
1. `CONFIG_PATH` environment variable
2. Default: `local.yaml` in the working directory

## Google Calendar Setup

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or use an existing one)
3. Enable the **Google Calendar API**: APIs & Services → Library → search "Google Calendar API" → Enable

### 2. Create a Service Account

1. Go to IAM & Admin → Service Accounts
2. Click "Create Service Account"
3. Give it a name (e.g., `calendar-manager`)
4. Skip the optional permissions steps (the service account does not need project-level roles)
5. Click "Done"

### 3. Generate a Key File

1. Click on the service account you just created
2. Go to the "Keys" tab
3. Add Key → Create new key → JSON
4. Save the downloaded JSON file as `credentials.json`

### 4. Share the Calendar with the Service Account

This is the critical step for least-privilege access. The service account only has access to calendars explicitly shared with it.

1. Open [Google Calendar](https://calendar.google.com/)
2. Find your target calendar in the left sidebar
3. Click the three dots → "Settings and sharing"
4. Under "Share with specific people or groups", click "Add people and groups"
5. Enter the service account email (found in the JSON key file as `client_email`, e.g., `calendar-manager@your-project.iam.gserviceaccount.com`)
6. Set permission to **"Make changes to events"**
7. Click "Send"

The service account now has permission to create, update, and delete events on this specific calendar only. It cannot access any other calendars or Google resources.

### 5. Get the Calendar ID and iCal URL

- **Calendar ID**: Settings and sharing → "Integrate calendar" → Calendar ID
- **iCal URL**: Settings and sharing → "Integrate calendar" → "Secret address in iCal format"

## Quick Start (local)

Prerequisites: Go 1.26+

1. Copy `local.example.yaml` to `local.yaml` and fill in your calendar details
2. Place your `credentials.json` in the project root
3. Run the server:
   ```sh
   go run ./cmd/server
   ```
4. Test:
   ```sh
   curl http://localhost:8080/probe
   curl http://localhost:8080/api/events
   ```

## Docker

```sh
docker-compose up --build
```

Requires `local.yaml` and `credentials.json` in the project root.

## Helm

```sh
helm upgrade --install calendar-manager charts/calendar-manager \
  --set config.calendar.calendarID="your-id@group.calendar.google.com" \
  --set config.calendar.icalURL="https://calendar.google.com/calendar/ical/.../basic.ics" \
  --set googleCredentials="$(cat credentials.json | base64 -w0)"
```

## Local K3D Development

```sh
make start-k3d
```

This creates a k3d cluster, builds and pushes the image to the local registry, and deploys via Helm.

## Make

Use `make help` to see available targets.

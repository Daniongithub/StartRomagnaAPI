# StartRomagnaAPI

# WORK IN PROGRESS, NOT FOR PRODUCTION!

Go backend parsing official GTFS and GTFS-RT streams from START Romagna

## .env example

```env
START_GTFS_ROOT=
START_GTFS_RT_ROOT=
WEB_AUTH_USER=
WEB_AUTH_PASSWORD=

DB_HOST=
DB_PORT=
DB_USERNAME=
DB_PASSWORD=

PORT=":5005"

# Array of allowed origins
ALLOWED_ORIGINS="https://example.com,https://example.com"

# Interval (in minutes) where future arrivals get displayed in /arrivals endpoint
ARRIVALS_LOAD_INTERVAL=90
```

## Features roadmap

### Foundation

- [X] Load config env
- [X] Healthcheck
- [X] HTTP server structure
- [X] CORS

### GTFS static importer

- [X] Fetch GTFS ZIP
- [X] Open ZIP in memory
- [X] Parse agency
- [X] Parse routes
- [X] Parse trips
- [X] Parse stops
- [X] Parse stop_times
- [X] Parse shapes
- [X] Parse calendar_dates
- [X] Validate relationships
- [X] Support RA
- [X] Support FC
- [X] Support RN

### Database

- [X] Define domain model
- [X] Design schema
- [X] Migrations
- [X] Store static GTFS
- [X] Update static GTFS safely
- [X] Check if DB is primary

### GTFS-RT importer

- [X] Fetch vehicle positions
- [X] Parse vehicle positions
- [X] Fetch trip updates
- [X] Parse trip updates
- [X] Fetch service alerts
- [X] Parse service alerts
- [X] Update realtime state
- [X] Support RA
- [X] Support FC
- [X] Support RN

### HTTP API

- [ ] GET /busesinservice
- [ ] GET /busesinservice/{basin}
- [X] GET /static/info
- [X] GET /static/trips/{basin}
- [ ] GET /static/trips/{trip_id}
- [X] GET /static/calendar_dates/{basin}
- [X] GET /static/routes/{basin}
- [X] GET /static/shapes/{basin}
- [X] GET /static/stop_times/{basin}
- [X] GET /static/stops/{basin}
- [ ] ...

### Backend integration

- [ ] Scan realtime and interface with ertpl_mezzi "fermi"

### Frontend integration

- [ ] Linee map
- [ ] Linee stop list
- [ ] Real-time vehicle map
- [ ] ...

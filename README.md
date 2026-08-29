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
- [ ] Update static GTFS safely
- [X] Check if DB is primary

### GTFS-RT importer

- [ ] Fetch vehicle positions
- [ ] Parse vehicle positions
- [X] Fetch trip updates
- [X] Parse trip updates
- [X] Fetch service alerts
- [X] Parse service alerts
- [ ] Update realtime state
- [X] Support RA
- [X] Support FC
- [X] Support RN

### HTTP API

- [ ] GET /busesinservice
- [ ] GET /busesinservice/{basin}
- [ ] GET /routes
- [ ] GET /routes/{route}
- [ ] GET /trips/{trip}
- [ ] GET /trips/{trip}/stops
- [ ] GET /trips/{trip}/shape
- [ ] GET /stops/{stop}
- [ ] ...

### Frontend integration

- [ ] Linee map
- [ ] Linee stop list
- [ ] Real-time vehicle map
- [ ] ...

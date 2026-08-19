# StartRomagnaAPI

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
```

## Features roadmap

### Foundation

- [X] Load config env
- [X] Healthcheck
- [ ] HTTP server structure
- [ ] CORS
- [ ] Logging

### GTFS static importer

- [ ] Fetch GTFS ZIP
- [ ] Open ZIP in memory
- [ ] Parse agency
- [ ] Parse routes
- [ ] Parse trips
- [ ] Parse stops
- [ ] Parse stop_times
- [ ] Parse shapes
- [ ] Parse calendar_dates
- [ ] Validate relationships
- [ ] Support RA
- [ ] Support FC
- [ ] Support RN

### Database

- [ ] Define domain model
- [ ] Design schema
- [ ] Migrations
- [ ] Store static GTFS
- [ ] Update static GTFS safely

### GTFS-RT importer

- [ ] Fetch vehicle positions
- [ ] Parse vehicle positions
- [ ] Fetch trip updates
- [ ] Parse trip updates
- [ ] Fetch service alerts
- [ ] Parse service alerts
- [ ] Update realtime state
- [ ] Support RA
- [ ] Support FC
- [ ] Support RN

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

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

### Data importer

- [X] Load config env
- [ ] Fetch static GTFS
- [ ] Store static GTFS
- [ ] Fetch GTFS-RT
- [ ] Store GTFS-RT

### HTTP endpoints

- [ ] CORS HTTP Headers
- [ ] Handlers setup
- [ ] GET /busesinservice
- [ ] GET /busesinservice/{id} separazione dei bacini di FC, RA, RN.
- [ ] To be defined...

# StartRomagnaAPI

Go backend parsing official GTFS and GTFS-RT streams from START Romagna.

Funzionamento live: <https://ertpl.pages.dev/start_menu/start>

Live operation: <https://ertpl.pages.dev/start_menu/start>

Progetto utilizzato per [Daniongithub/ER-TPL](https://github.com/Daniongithub/ER-TPL).

Project used for [Daniongithub/ER-TPL](https://github.com/Daniongithub/ER-TPL).

## Informazioni - Notice

Il progetto cerca di rendere facile la gestione dei feed GTFS e GTFS-RT forniti da START Romagna.
È necessario chiedere gli endpoint ufficiali a START, siccome nascosti e protetti da Basic Authentication.
Il progetto non può essere riutilizzato per altri scopi, siccome il codice è troppo specifico, sarebbe impossibile riusarlo.

> [!IMPORTANT]
> I bacini sono: FC - RA - RN. Provincie di Forlì-Cesena, Ravenna e Rimini.

---

This project tries to simplify the management of GTFS and GTFS-RT feeds given by START Romagna.
It's necessary to ask them their official endpoints, since they're hidden and protected with Basic Authentication.
This project can't be reutilised, because the code is too specific, it'd be impossible to use it elsewhere.

> [!IMPORTANT]
> The basins are: FC - RA - RN. Provinces of Forlì-Cesena, Ravenna e Rimini in Italy.

## Endpoints

- `GET /health` -> Ritorna HTTP 200, utile per un healthcheck. | Returns HTTP 200, useful for an healthcheck.
- `GET /rss/feed` -> Feed RSS di Infobus Start, in formato JSON. | Infobus Start RSS Feed, in JSON format.
- `GET /arrivals/{stopcode}` -> Bus previsti in arrivo per la fermata specificata. | Returns the list of buses predicted to be arriving in the specified bus stop.
- `GET /busesinservice` -> Tutti i bus in servizio di START, in tempo reale, con informazioni sui mezzi. | Returns all the buses currently in service, with some vehicle information.
- `GET /activevehicles` -> Verrà deprecato. | To be deprecated.
- `GET /linelist/{basin}` -> ???
- `GET /nextstops/{tripid}` -> Ritorna le prossime fermate di una corsa in svolgimento. | Returns the next stops of an ongoing trip.
- `GET /vehiclepositions` -> Tutte le posizioni dei veicoli in tempo reale. | All vehicles' positions.
- `GET /vehiclepositions/{basin}` -> `/vehiclepositions` di un solo bacino. | `/vehiclepositions` of a single basin.
- `GET /vehicleposition/{vehicleid}` -> `/vehiclepositions` di un solo veicolo. | `/vehiclepositions` of a single vehicle.
- `GET /shape/{basin}/{shapeId}` -> Shape del percorso, dato il bacino e l'id univoco. | Shape of the trip, given the basin and its unique id.
- `GET /vehicleinfo/{vehicle}` -> Informazioni su un singolo veicolo. | Return information of a signle vehicle.
- `GET /stopsinfo/{basin}/{stopcode}` -> Ritorna tutte le linee che passano in una fermata, dato il bacino ed il codice fermata. | Returns all the lines that stop in the specified bus stop, given the basin and its id.
- `GET /static/info` -> Range di date in cui sono validi i dati del GTFS statico. | Date range where static GTFS data is valid.
- `GET /static/trips/{basin}` -> Lista di tutti i trip di un bacino. | Returns the list of all the trips of a basin.
- `GET /static/routes/{basin}` -> Lista di tutte le linee di un bacino. | Returns the list of all the routes of a basin.
- `GET /static/shapes/{basin}` -> Lista di tutte le shape di un bacino. | Returns the list of all the shapes of a basin.
- `GET /static/stops/{basin}` -> Lista di tutte le fermate di un bacino. | Returns the list of all the bus stops of a basin.

### Da implementare - To be implemented

- `GET /timetable/{routeid}`
- `/events`

## Note sul funzionamento - Notes on operation

Vengono eseguite delle task ogni 20 secondi per il realtime, alle 04:00 si aggiorna il GTFS statico.

Usiamo un singolo database MariaDB, con 3 DB dentro: `start_gtfs_static`, `start_gtfs_rt` e `ertpl_mezzi` (il DB con tutte le informazioni sui veicoli).
Chiedici gli schemas, se servono.

Some tasks are executed every 20 seconds for the realtime, at 04:00 the static GTFS gets updated.

We use a single MariaDB database, with 3 DB inside: `start_gtfs_static`, `start_gtfs_rt` and `ertpl_mezzi` (our vehicles' information DB).
Ask us the schemas, if needed.

## ENV Config

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
# Max visualized delay before trip gets deleted from /arrivals endpoint
ARRIVALS_DELAY_BUFFER=90
```

> [!WARNING]
> Tutte queste env vars sono RICHIESTE. Se non fornite, il programma uscirà. | All of these env vars are MANDATORY. If those are without values, the program will exit.

## Crediti - Credits

Ringraziamo di cuore il Centro Elaborazione Dati di START Romagna per averci fornito l'accesso ai dati GTFS e realizzato il nostro sogno.

Un grande GRAZIE va a [@Leocraft1](https://github.com/Leocraft1) perché senza di lui questo progetto sarebbe nato male e "storto".

---

Many thanks to START Romagna's IT team.

A big THANK YOU goes to [@Leocraft1](https://github.com/Leocraft1) because without him this project would have got off to a bad start.

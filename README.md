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

- `GET /health` -> Stato del database in JSON; HTTP 500 se la verifica fallisce. | Database status in JSON; HTTP 500 when the check fails.
- `GET /rss/feed` -> Feed RSS di Infobus Start, in formato JSON. | Infobus Start RSS Feed, in JSON format.
- `GET /arrivals/{stopcode}` -> Bus previsti in arrivo per la fermata specificata. | Returns the list of buses predicted to be arriving in the specified bus stop.
- `GET /busesinservice` -> Tutti i bus in servizio di START, in tempo reale, con informazioni sui mezzi. | Returns all the buses currently in service, with some vehicle information.
- `GET /nextstops/{tripid}` -> Ritorna le prossime fermate di una corsa in svolgimento. | Returns the next stops of an ongoing trip.
- `GET /timetable/{basin}/{routeid}/{directionid}` -> Ritorna tutte le corse della linea selezionata oggi. DirectionId è 0 o 1 | Returns all trips for the desired route today. DirectionId is 0 or 1.
- `GET /vehiclepositions` -> Tutte le posizioni dei veicoli in tempo reale. | All vehicles' positions.
- `GET /vehiclepositions/{basin}` -> `/vehiclepositions` di un solo bacino. | `/vehiclepositions` of a single basin.
- `GET /vehicleposition/{vehicleid}` -> `/vehiclepositions` di un solo veicolo. | `/vehiclepositions` of a single vehicle.
- `GET /shape/{basin}/{shapeId}` -> Shape del percorso, dato il bacino e l'id univoco. | Shape of the trip, given the basin and its unique id.
- `GET /shapeslist/{basin}/{routeId}` -> Lista di shape_id dato il bacino e il route_id. | List of all shape_id given the basin and route_id.
- `GET /vehicleinfo/{vehicle}` -> Informazioni su un singolo veicolo. | Return information of a signle vehicle.
- `GET /stopsinfo/{basin}/{stopcode}` -> Ritorna tutte le linee che passano in una fermata, dato il bacino ed il codice fermata. | Returns all the lines that stop in the specified bus stop, given the basin and its id.
- `GET /announcements` -> Annunci | Announcements
- `GET /events` -> Stream SSE delle importazioni realtime completate. | SSE stream of completed realtime imports.
- `GET /static/info` -> Range di date in cui sono validi i dati del GTFS statico. | Date range where static GTFS data is valid.
- `GET /static/trips/{basin}` -> Lista di tutti i trip di un bacino. | Returns the list of all the trips of a basin.
- `GET /static/routes/{basin}` -> Lista di tutte le linee di un bacino. | Returns the list of all the routes of a basin.
- `GET /static/shapes/{basin}` -> Lista di tutte le shape di un bacino. | Returns the list of all the shapes of a basin.
- `GET /static/stops/{basin}` -> Lista di tutte le fermate di un bacino. | Returns the list of all the bus stops of a basin.
- `/panel/announcements/` -> Pannello annunci protetto da OIDC; disponibile solo con `ENABLE_AUTH=true`.
- `GET /api/announcements` -> Alias pubblico di `GET /announcements`.

## Note sul funzionamento - Notes on operation

Trip update e posizioni vengono aggiornati ogni 15 secondi; gli alert ogni minuto. Alle 04:00 si aggiorna il GTFS statico. Lo scheduler usa il fuso `Europe/Rome`.

Usiamo un singolo database MariaDB, con 3 DB dentro: `start_gtfs_static`, `start_gtfs_rt` e `ertpl_mezzi` (il DB con tutte le informazioni sui veicoli).
Chiedici gli schemas, se servono.

Trip updates and vehicle positions refresh every 15 seconds; alerts refresh every minute. Static GTFS refreshes at 04:00. The scheduler uses the `Europe/Rome` timezone.

We use a single MariaDB database, with 3 DB inside: `start_gtfs_static`, `start_gtfs_rt` and `ertpl_mezzi` (our vehicles information DB).
Ask us the schemas, if needed.

Gli import GTFS sono transazionali: un errore nel download, nella decodifica o nella scrittura conserva lo stato precedente. Le tabelle devono usare un motore transazionale, come InnoDB. Le notifiche SSE vengono emesse dopo il commit. L'RSS continua a essere scaricato direttamente dal sito.

All responses include `Access-Control-Allow-Origin: *`, including errors and OPTIONS responses. No origin allowlist is required.

## Command-line execution

```bash
go build -o startromagnaapi .
./startromagnaapi --update-static-gtfs
```

Il flag `--update-static-gtfs` esegue una sola volta l'importazione statica dei tre bacini, usando la stessa operazione dello scheduler delle 04:00, poi termina. Apre soltanto il pool del database statico; non avvia server HTTP, SSE, scheduler, realtime o discovery OIDC. Senza flag il binario continua a funzionare come servizio. `--help` mostra le opzioni senza caricare la configurazione.

Eseguire il comando dalla directory che contiene `.env`. Per questa modalità servono soltanto `START_GTFS_ROOT`, `WEB_AUTH_USER`, `WEB_AUTH_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_USERNAME` e `DB_PASSWORD`.

Un lock MariaDB condiviso protegge sia il comando sia il job dello scheduler: se un'importazione statica è già in corso sullo stesso server DB, il secondo tentativo termina senza avviare download o scritture. Questa protezione richiede che anche il servizio utilizzi la versione aggiornata del binario. La sostituzione dei dati rimane transazionale.

Exit codes: `0` for success, `1` for an import/startup failure or an already running import, `2` for invalid command-line arguments. Without flags, the binary starts the normal API service.

## ENV Config

```env
START_GTFS_ROOT=
START_GTFS_RT_ROOT=
WEB_AUTH_USER=
WEB_AUTH_PASSWORD=

OIDC_ISSUER_URL=
OIDC_CLIENT_ID=
OIDC_ALLOWED_GROUPS=
ENABLE_AUTH=false

DB_HOST=
DB_PORT=
DB_USERNAME=
DB_DOMAIN=
DB_PASSWORD=
# db orchestrating config
DB_ORCH_USERNAME=
DB_ORCH_PASSWORD=
REPLICA_USERNAME=
REPLICA_PASSWORD=
ARBITER_SECRET=

PORT=":5005"

# Interval (in minutes) where future arrivals get displayed in /arrivals endpoint
ARRIVALS_LOAD_INTERVAL=90
# Max visualized delay before trip gets deleted from /arrivals endpoint
ARRIVALS_DELAY_BUFFER=90
```

Con `ENABLE_AUTH=true`, il redirect URI del client OIDC deve essere
`https://<host>/panel/announcements/auth/callback` (o `http` per ambienti
senza TLS). `OIDC_ALLOWED_GROUPS` accetta gruppi separati da virgole. Con
`ENABLE_AUTH=false` il pannello non viene registrato; gli altri endpoint API
restano disponibili.

La pagina iniziale e quella di login, insieme agli eventuali asset, sono in
`internal/panels/announcements/static`. Per collegare handler Go al pannello,
registrarli con `Panel.RegisterAPI` prima di montare `Panel.Mux()`; i relativi
endpoint sono protetti dalla sessione OIDC.

> [!WARNING]
> Tutte queste env vars sono RICHIESTE. Se non fornite, il programma uscirà. | All of these env vars are MANDATORY. If those are without values, the program will exit.

## Crediti - Credits

Ringraziamo di cuore il Centro Elaborazione Dati di START Romagna per averci fornito l'accesso ai dati GTFS e realizzato il nostro sogno.

Un grande GRAZIE va a [@Leocraft1](https://github.com/Leocraft1) perché senza di lui questo progetto sarebbe nato male e "storto".

---

Many thanks to START Romagna's IT team.

A big THANK YOU goes to [@Leocraft1](https://github.com/Leocraft1) because without him this project would have got off to a bad start.

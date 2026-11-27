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
- `GET /timetable/{basin}/{routeid}/{directionid}` -> Ritorna tutte le corse della linea selezionata oggi e la lista globale dei passaggi in fermata. DirectionId è 0 o 1. | Returns today's trips and a global list of stop passages. DirectionId is 0 or 1.
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

The timetable response is an object with `stops` and `trips`. The previous top-level trip array is now available under `trips`.

`stops` contains every passage of every returned trip, including repeated visits and stops shared by different trips. Each entry includes `stop_name`, `stop_code`, `stop_id`, `trip_id`, `stop_sequence`, `arrival_time` and `departure_time`. Entries are sorted by `stop_sequence`, then arrival time (including GTFS hours above 24:00), then `trip_id` for ties. Stops marked `is_dummy = true` are excluded from both the global list and each trip's `stop_times`. Empty lists are returned as `[]`.

## Note sul funzionamento - Notes on operation

Trip update e posizioni vengono aggiornati ogni 15 secondi; gli alert ogni minuto. Alle 04:00 si aggiorna il GTFS statico. Lo scheduler usa il fuso `Europe/Rome`.

Usiamo un singolo database MariaDB, con 3 DB dentro: `start_gtfs_static`, `start_gtfs_rt` e `ertpl_mezzi` (il DB con tutte le informazioni sui veicoli).
Chiedici gli schemas, se servono.

Trip updates and vehicle positions refresh every 15 seconds; alerts refresh every minute. Static GTFS refreshes at 04:00. The scheduler uses the `Europe/Rome` timezone.

We use a single MariaDB database, with 3 DB inside: `start_gtfs_static`, `start_gtfs_rt` and `ertpl_mezzi` (our vehicles information DB).
Ask us the schemas, if needed.

Gli import statici aggiornano ogni bacino in una transazione separata. Se un download, la validazione o una scrittura falliscono, il bacino interessato conserva i dati precedenti e l'importazione prosegue con gli altri bacini. Un lock MariaDB evita aggiornamenti statici concorrenti sullo stesso database. Le tabelle devono usare un motore transazionale, come InnoDB. Le notifiche SSE vengono emesse dopo il commit. L'RSS continua a essere scaricato direttamente dal sito.

All responses include `Access-Control-Allow-Origin: *`, including errors and OPTIONS responses. No origin allowlist is required.

## Command-line execution

```bash
go build -o startromagnaapi .
./startromagnaapi
```

Eseguire il comando dalla directory che contiene `.env`.

## ENV Config

```env
START_GTFS_FTP=
FTP_USERNAME=
FTP_PASSWORD=

# OAuth configuration
OIDC_ISSUER_URL=
OIDC_CLIENT_ID=
OIDC_ALLOWED_GROUPS=
ENABLE_AUTH=false # enable auth-related endpoints only if auth provider is set

DB_HOST=
DB_PORT=
DB_USERNAME=
DB_PASSWORD=

IS_PRIMARY=true # sync static GTFS and vehicle status on DB

PORT=":5005"

# Interval (in minutes) where future arrivals get displayed in /arrivals endpoint
ARRIVALS_LOAD_INTERVAL=90
# Max visualized delay before trip gets deleted from /arrivals endpoint
ARRIVALS_DELAY_BUFFER=90

# Optional Telegram notifications
TELEGRAM_ENABLED=false
TELEGRAM_BOT_TOKEN=
# Comma-separated numeric chat IDs
TELEGRAM_CHAT_IDS=
# Optional name for this instance. Defaults to the system hostname.
TELEGRAM_HOST_NAME=
# Writable, persistent directory. Defaults to .telegram-state next to .env.
TELEGRAM_STATE_DIR=
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
> Le variabili `OIDC_*` sono richieste quando `ENABLE_AUTH=true`. Il token e gli ID delle chat Telegram sono richiesti quando `TELEGRAM_ENABLED=true`. Le variabili di configurazione GTFS, database, porta, ruolo e intervalli sono obbligatorie.

### Notifiche Telegram

Le notifiche sono facoltative. Con `TELEGRAM_ENABLED=true`, ogni istanza segnala gli errori operativi dei job GTFS statici e realtime, del database, del feed RSS e dell'API. Per gli import statici l'avviso indica anche il bacino e la fase fallita. Gli errori ravvicinati vengono raggruppati, con un avviso per operazione ogni 30 minuti per host. Il riepilogo arriva via Telegram e il dettaglio tecnico è nei log.

Il messaggio di prova viene inviato alla prima associazione tra bot, host e chat. Ai riavvii l'istanza usa lo stato salvato in `.telegram-state`, accanto a `.env`. Per i deploy che sostituiscono la directory di lavoro, imposta `TELEGRAM_STATE_DIR` su una directory persistente e scrivibile dall'utente del servizio. Una nuova chat riceve il test al primo avvio con la configurazione aggiornata.

Per gli errori fatali, il processo tenta la consegna prima di terminare, con un limite di 5 secondi. Gli errori nella lettura di `.env` possono essere notificati se la configurazione Telegram valida è già presente nell'ambiente del processo.

```env
TELEGRAM_ENABLED=true
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_IDS=123456789,987654321
TELEGRAM_HOST_NAME=server-daniele
```

Usa lo stesso token su entrambi gli host e assegna a ciascuno un `TELEGRAM_HOST_NAME` diverso. Se omesso, viene usato il nome di sistema. Ogni host invia ai destinatari elencati nella sua configurazione. Puoi usare entrambi gli ID privati o l'ID di un gruppo che contiene il bot. Tratta il token come una password.

Per trovare gli ID privati, ogni destinatario deve aprire il bot e inviargli `/start`. Leggi poi `message.chat.id` per ciascun messaggio con il metodo [getUpdates della Bot API](https://core.telegram.org/bots/api#getupdates). Per un gruppo, aggiungi il bot e invia un comando rivolto al bot nel gruppo.

## Crediti - Credits

Ringraziamo di cuore il Centro Elaborazione Dati di START Romagna per averci fornito l'accesso ai dati GTFS e realizzato il nostro sogno.

Un grande GRAZIE va a [@Leocraft1](https://github.com/Leocraft1) perché senza di lui questo progetto sarebbe nato male e "storto".

---

Many thanks to START Romagna's IT team.

A big THANK YOU goes to [@Leocraft1](https://github.com/Leocraft1) because without him this project would have got off to a bad start.

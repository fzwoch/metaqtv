# MetaQTV

> MetaQtv is a web API that serves a collection of QuakeWorld Servers from multiple QuakeWorld masters.

## Usage

```sh
metaqtv [-interval INTERVAL] [-port PORT]
```

| arg        | type  | description                | default | 
|------------|-------|----------------------------|---------|
| `interval` | `int` | Update interval in seconds | `60`    | 
| `port`     | `int` | HTTP listen port           | `3000`  |

## API endpoints

* `/api/servers` - "Normal" Quake servers
* `/api/proxies` - Proxies
* `/api/qtv` - QTV servers
* `/api/qtv_to_servers` - Map of QTV stream URLs to server addresses
* `/api/server_to_qtv` - Map of server addresses to QTV stream URLs

## Config

### Master servers

The QuakeWorld master servers to query for servers.

**Example**
`master_servers.json`

```json
[
  "master.quakeworld.nu:27000",
  "master.quakeservers.net:27000",
  "qwmaster.ocrana.de:27000",
  "qwmaster.fodquake.net:27000"
]
```

## Build

```sh
$ go build
```

## Credits

* eb
* Tuna

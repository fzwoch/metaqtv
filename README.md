# MetaQTV

> Web API serving QuakeWorld server info

## Usage

```sh
metaqtv [-master INTERVAL] [-server INTERVAL] [-active INTERVAL] [-port PORT]
```

| arg      | type  | description                   | default | 
|----------|-------|-------------------------------|---------|
| `port`   | `int` | HTTP listen port              | `3000`  |
| `master` | `int` | Master server update interval | `600`   |
| `server` | `int` | Server update interval        | `30`    |
| `active` | `int` | Active server update interval | `3`     |

## API endpoints

| URL                 | description                            |  
|---------------------|----------------------------------------|
| `/v2/servers`       | Mvdsv servers                          |  
| `/v2/qwfwd`         | Qwfwd servers (proxies)                |  
| `/v2/qtv`           | QTV servers                            |  
| `/v2/qtv_to_server` | Map of QTV streams to server addresses |  
| `/v2/server_to_qtv` | Map of server addresses to QTV streams |

### Query params

| URL                     | description                                                |
|-------------------------|------------------------------------------------------------|
| `?has_client=xantom`    | Servers where `xantom` is connected as player or spectator |
| `?has_player=xantom`    | Servers where `xantom` is connected as player              |
| `?has_spectator=xantom` | Servers where `xantom` is connected as spectator           |
|                         |                                                            |
| `?mode=ffa`             | Servers where `Mode` is `ffa`                              |
| `?mode=2on2&mode=4on4`  | Servers where `Mode` is `2on2` or `4on4`                   |
| `?status=Started`       | Servers where `Status` is `Started`                        |
| `?settings.map=dm3`     | Servers where `map` is `dm3`                               |
| `?player_count=gte:3`   | Servers with at least 3 players                            |
| `?sort_by=address`      | Sort by server `address`                                   |
| `?sort_order=desc`      | Sort in `descending` order                                 |

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
* XantoM

# MetaQTV

> Web API serving QuakeWorld server info

## Usage

```sh
metaqtv [-interval INTERVAL] [-port PORT]
```

| arg        | type  | description                | default | 
|------------|-------|----------------------------|---------|
| `interval` | `int` | Update interval in seconds | `60`    | 
| `port`     | `int` | HTTP listen port           | `3000`  |

## API endpoints

| URL               | description                            |  
|-------------------|----------------------------------------|
| `/servers`        | "Normal" Quake servers                 |  
| `/proxies`        | Proxies                                |  
| `/qtv`            | QTV servers                            |  
| `/qtv_to_servers` | Map of QTV streams to server addresses |  
| `/server_to_qtv`  | Map of server addresses to QTV streams |  

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

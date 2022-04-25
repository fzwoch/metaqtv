# MetaQTV

> An improved Go re-implementation of [MetaQTV](https://github.com/eb/metaqtv/).

## Usage

```sh
metaqtv [-interval INTERVAL] [-port PORT]
```

| arg        | type  | description                | default | 
|------------|-------|----------------------------|---------|
| `interval` | `int` | Update interval in seconds | `60`    | 
| `port`     | `int` | HTTP listen port           | `3000`  | 

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

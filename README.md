# MetaQTV

An improved Go re-implementation of [MetaQTV](https://github.com/eb/metaqtv/).

## Usage

```
Usage of metaqtv:
  -config string
    	QTV server config file (default "metaqtv.json")
  -interval int
    	Update interval in seconds (default 60)
  -port int
    	HTTP listen port (default 3000)
  -timeout int
    	RSS request timeout in seconds (default 5)
```

## Config

Application reads `metaqtv.json` from the current working directory. This config file lists Quake master servers which are queried.

```json
[
    {
        "hostname": "qwmaster.ocrana.de",
        "port": 27000
    },
    [..]
]
```

## Build

```
$ go build
```

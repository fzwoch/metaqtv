# MetaQTV

An improved Go re-implementation of [MetaQTV](https://github.com/eb/metaqtv/).

## Usage

```
Usage of metaqtv:
  -config string
    	Master server config file (default "metaqtv.json")
  -interval int
    	Update interval in seconds (default 60)
  -port int
    	HTTP listen port (default 3000)
  -retry int
    	UDP retry count (default 5)
  -timeout int
    	UDP timeout in milliseconds (default 500)
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

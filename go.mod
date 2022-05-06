module metaqtv

go 1.18

require (
	github.com/victorspringer/http-cache v0.0.0-20220131145941-ef3624e6666f
	github.com/vikpe/masterstat v0.1.6
	github.com/vikpe/serverstat v0.1.7
)

require github.com/vikpe/udpclient v0.1.2 // indirect


replace github.com/vikpe/serverstat => ../serverstat

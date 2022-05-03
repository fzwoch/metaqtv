module metaqtv

go 1.18

require (
	github.com/victorspringer/http-cache v0.0.0-20220131145941-ef3624e6666f
	github.com/vikpe/qw-masterstat v0.1.3
	github.com/vikpe/qw-serverstat v0.1.3
)

require github.com/vikpe/udpclient v0.1.0 // indirect

replace github.com/vikpe/qw-serverstat v0.1.3 => ../qw-serverstat

module github.com/Magic-Pod/magicpod-api-client

go 1.24.2

require (
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000
	github.com/urfave/cli v1.22.5
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
)

replace github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0

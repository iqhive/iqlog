module github.com/iqhive/iqlog/bench

go 1.23

replace github.com/iqhive/iqlog => ../

require github.com/iqhive/iqlog v0.0.0

require (
	github.com/phuslu/log v1.0.128
	github.com/rs/zerolog v1.35.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

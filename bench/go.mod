module github.com/iqhive/iqlog/bench

go 1.26.0

replace github.com/iqhive/iqlog => ../

require github.com/iqhive/iqlog v1.0.26

require bitbucket.org/iqhive/iqlog/v3 v3.0.0

replace bitbucket.org/iqhive/iqlog/v3 => ./legacyshim

require (
	github.com/phuslu/log v1.0.133
	github.com/rs/zerolog v1.35.1
)

require (
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
)

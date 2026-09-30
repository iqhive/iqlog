module github.com/iqhive/iqlog/examples/hero

go 1.26.0

require (
	github.com/iqhive/banner v1.0.0
	github.com/iqhive/iqlog v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
)

replace github.com/iqhive/iqlog => ../..

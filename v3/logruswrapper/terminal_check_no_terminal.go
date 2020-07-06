// +build js nacl plan9

package logruswrapper

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return false
}

// +build appengine

package logruswrapper

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return true
}

// Useful functions.
package utils

import (
	"io"
	"os"

	"github.com/PatronC2/Patron/lib/logger"
)

/*
	Utiliy to close an object that implements the `io.Closer` interface. Used
	over `object.Close()` for automatic error checking.
*/
func Close(closable io.Closer) {
	if err := closable.Close(); err != nil {
		logger.LogError(err)
		os.Exit(logger.ERR_CLOSABLE)
	}
}

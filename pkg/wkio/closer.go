package wkio

import (
	"io"
	"log"
)

// SafeClose close an io.Closer in a more pedantic way to avoid lint warnings.
func SafeClose(c io.Closer) {
	if c != nil {
		err := c.Close()
		if err != nil {
			log.Println(err)
		}
	}
}

// Close is a more complex version of SafeClose
func Close(c io.Closer, f func(string, ...interface{}), msg string) {
	if c != nil {
		err := c.Close()
		if err != nil && f != nil {
			f(msg, err)
		}
	}
}

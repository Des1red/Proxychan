package commands

import (
	"fmt"
	"os"

	"proxychan/internal/models"
)

func fatal(err error) {
	msg, code := models.FormatForUser(err)
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

package main

import (
	"fmt"

	"github.com/master-g/i18n/internal/model"
)

func main() {
	raw := `this is a complicate string with ', and & this, like ", >_< @me`
	escaped := model.EscapeString(raw)
	fmt.Println(escaped)
}

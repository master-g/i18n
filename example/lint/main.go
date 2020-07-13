package main

import (
	"fmt"

	"github.com/master-g/i18n/internal/model"
)

func main() {
	l := model.WithDefaultLinters()
	results := l("en", "key", "this $1%t is a good string s%, ％t")
	for _, r := range results {
		fmt.Println(r.Desc)
	}
}

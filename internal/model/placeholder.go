package model

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type placeholderContext struct {
	index     int
	r         rune
	origin    string
	replaceTo string
}

func AutoPlaceholder(raw string) string {
	const stateSearchingHead = 0
	const stateHeadFound = 1
	const stateFirstLetterCaptured = 2

	state := stateSearchingHead
	var char rune

	var contextList []*placeholderContext

	// 1. find all %AA, %BB, ..., %ZZ
	for _, r := range raw {
		switch state {
		case stateSearchingHead:
			if r == '%' {
				state = stateHeadFound
			}
		case stateHeadFound:
			if unicode.IsUpper(r) {
				char = r
				state = stateFirstLetterCaptured
			} else {
				state = stateSearchingHead
			}
		case stateFirstLetterCaptured:
			if r == char {
				contextList = append(contextList, &placeholderContext{
					index:     0,
					r:         r,
					origin:    string([]rune{'%', r, r}),
					replaceTo: "",
				})
			}
			state = stateSearchingHead
		}
	}

	// 2. index and replace
	after := raw

	sort.Slice(contextList, func(i, j int) bool {
		return int(contextList[i].r) < int(contextList[j].r)
	})

	r2i := make(map[rune]int)
	index := 1

	for _, ctx := range contextList {
		if i, ok := r2i[ctx.r]; !ok {
			r2i[ctx.r] = index
			ctx.index = index
			index++
		} else {
			ctx.index = i
		}
		ctx.replaceTo = fmt.Sprintf("%%%d$s", ctx.index)

		after = strings.ReplaceAll(after, ctx.origin, ctx.replaceTo)
	}

	return after
}

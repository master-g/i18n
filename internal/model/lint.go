package model

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type LintResult struct {
	Language string `json:"language"`
	Key      string `json:"key"`
	Desc     string `json:"desc"`
}

func (lr *LintResult) String() string {
	raw, err := json.Marshal(lr)
	if err != nil {
		return err.Error()
	}

	return string(raw)
}

type Linter func(lang, key, raw string) []*LintResult

func (s *SourceFile) Lint(linters ...Linter) (result []*LintResult) {
	if s == nil || len(s.Languages) == 0 || len(linters) == 0 {
		return
	}

	for lang, kvs := range s.Languages {
		for key, str := range kvs.KVS {
			for _, linter := range linters {
				lintResultOfSingleLine := linter(lang, key, str)
				result = append(result, lintResultOfSingleLine...)
			}
		}
	}

	return
}

// builtin linters

func WithDefaultLinters() Linter {
	return func(lang, key, raw string) []*LintResult {
		builtinLinters := []Linter{
			lintFormatSpecifiers,
			// ADD OTHER LINTER HERE
		}

		var result []*LintResult
		for _, linter := range builtinLinters {
			r := linter(lang, key, raw)
			if len(r) > 0 {
				result = append(result, r...)
			}
		}

		return result
	}
}

func lintFormatSpecifiers(lang, key, raw string) []*LintResult {
	var result []*LintResult

	reg1 := regexp.MustCompile(`[a-zA-Z]%`)
	reg2 := regexp.MustCompile(`\$\d%[a-zA-Z]`)

	reg3 := regexp.MustCompile(`％[a-zA-Z]`)
	reg4 := regexp.MustCompile(`[a-zA-Z]％`)
	reg5 := regexp.MustCompile(`％\d$[a-zA-Z]`)
	reg6 := regexp.MustCompile(`\$\d％[a-zA-Z]`)

	regList := []*regexp.Regexp{reg1, reg2, reg3, reg4, reg5, reg6}

	for _, reg := range regList {
		indices := reg.FindAllIndex([]byte(raw), -1)
		for _, index := range indices {
			r := &LintResult{
				Language: lang,
				Key:      key,
				Desc:     fmt.Sprintf("invalid format specifier '%v' found in lang:%v, key:%v, at pos:%v", raw[index[0]:index[1]], lang, key, index[0]),
			}
			result = append(result, r)
		}
	}

	return result
}

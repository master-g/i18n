package parser

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/master-g/i18n/internal/model"
)

type CollisionResolver func(path, key, pre, cur string) string

func LoadCSV(p string, collisionResolver CollisionResolver) (ret *model.SourceFile, err error) {
	var csvFile *os.File
	if !filepath.IsAbs(p) {
		p, err = filepath.Abs(p)
		if err != nil {
			return
		}
	}
	csvFile, err = os.Open(p)
	if err != nil {
		return
	}
	defer func() {
		err2 := csvFile.Close()
		if err == nil {
			err = err2
		}
	}()
	csvReader := csv.NewReader(bufio.NewReader(csvFile))

	// index to language
	index2lang := make(map[int]string)

	tmp := &model.SourceFile{
		Type:      model.SourceFileTypeCSV,
		AbsPath:   p,
		Languages: make(map[string]*model.LanguageKVS),
	}

	for {
		var records []string
		records, err = csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}
		if len(index2lang) == 0 {
			for i, lang := range records {
				if i > 0 && lang != "" {
					index2lang[i] = lang
				}
			}
		} else {
			var strKey string
			for i, str := range records {
				if str == "" {
					continue
				}

				if i == 0 {
					// key
					strKey = strings.TrimSpace(str)
					if strKey == "" {
						err = errors.New("empty key found in source file")
						return
					}
				} else if lang, ok := index2lang[i]; ok {
					if tmp.Languages[lang] == nil {
						tmp.Languages[lang] = &model.LanguageKVS{
							Language: lang,
							KVS:      make(map[string]string),
						}
					}
					newValue := strings.TrimSpace(str)
					oldEntry, collision := tmp.Languages[lang].KVS[strKey]
					if collision && strings.Compare(oldEntry, newValue) != 0 {
						if collisionResolver != nil {
							newValue = collisionResolver(p, strKey, oldEntry, newValue)
						}
					}
					tmp.Languages[lang].KVS[strKey] = newValue
				}
			}
		}
	}

	ret = tmp

	return
}

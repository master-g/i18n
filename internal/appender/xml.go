package appender

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/master-g/i18n/internal/model"
	"github.com/master-g/i18n/pkg/wkfs"
)

type xmlLine struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Pos   int    `json:"pos"`
	Lines int    `json:"lines"`
}

type CollisionResolver func(file string, pos int, key, old, newer string) string

func AppendToXML(data map[string]string, output string, resolver CollisionResolver, dry bool) (keyCollisions, keyAppended int, err error) {
	var lines []string
	lines, err = wkfs.ReadAllLines(output)
	if err != nil {
		return
	}

	newFileLines := make([]string, 0, len(lines))

	multiLineStart := 0
	multiLineEnd := 0
	multiLine := false
	buf := &strings.Builder{}
	oldSet := make(map[string]*xmlLine)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		found := false

		if multiLine {
			buf.WriteString(trimmed)
			if strings.HasPrefix(trimmed, "<string name=") {
				err = fmt.Errorf("invalid xml, line: %d", i)
				return
			} else if strings.HasSuffix(trimmed, "</string>") {
				trimmed = buf.String()
				buf.Reset()
				found = true
				multiLineEnd = i
			}
		} else if strings.HasPrefix(trimmed, "<string name=") {
			if strings.HasSuffix(trimmed, "</string>") {
				found = true
			} else {
				multiLineStart = i
				multiLine = true
				buf.WriteString(trimmed)
			}
		}

		if found {
			m := &model.StringXMLItem{}
			err = xml.Unmarshal([]byte(trimmed), m)
			if err != nil {
				return
			}

			pos := i
			numOfLines := 1
			if multiLine {
				pos = multiLineStart
				numOfLines = multiLineEnd - multiLineStart + 1
			}

			oldSet[m.Name] = &xmlLine{
				Key:   m.Name,
				Value: m.Value,
				Pos:   pos,
				Lines: numOfLines,
			}

			multiLine = false
			multiLineStart = 0
			multiLineEnd = 0
		} else if strings.HasPrefix(trimmed, "</resources>") {
			break
		}
		newFileLines = append(newFileLines, line)
	}

	appendFlag := false
	lastLine := strings.TrimSpace(newFileLines[len(newFileLines)-1])
	if lastLine == "" {
		appendFlag = true
	}

	sortedKeys := make([]string, 0, len(data))
	for key := range data {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := data[key]
		newEntry := &model.StringXMLItem{
			Name:  key,
			Value: value,
		}
		var newBytes []byte
		newBytes, err = xml.MarshalIndent(newEntry, "    ", "")
		if err != nil {
			return
		}
		newLine := string(newBytes)

		if oldEntry, ok := oldSet[key]; ok {
			if value != strings.TrimSpace(oldEntry.Value) {
				keyCollisions++
				if resolver != nil {
					value = resolver(output, oldEntry.Pos, key, oldEntry.Value, value)
				} else {
					value = oldEntry.Value
				}
				m := &model.StringXMLItem{
					Name:  key,
					Value: value,
				}

				// replace
				var replaceData []byte
				replaceData, err = xml.MarshalIndent(m, "    ", "")
				if err != nil {
					return
				}
				newFileLines[oldEntry.Pos] = string(replaceData)
				if oldEntry.Lines > 1 {
					newFileLines = append(newFileLines[:oldEntry.Pos+1], newFileLines[oldEntry.Pos+oldEntry.Lines:]...)
				}
			}
			continue
		}

		keyAppended++
		if !appendFlag {
			appendFlag = true
			newFileLines = append(newFileLines, "\n")
		}
		newFileLines = append(newFileLines, newLine)
	}

	sb := &strings.Builder{}
	for _, l := range newFileLines {
		sb.WriteString(l)
		if l != "\n" {
			sb.WriteRune('\n')
		}
	}
	sb.WriteString("</resources>")

	if !dry {
		err = ioutil.WriteFile(output, []byte(sb.String()), 0644)
	}

	return
}

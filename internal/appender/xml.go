package appender

import (
	"encoding/xml"
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
}

type CollisionResolver func(file string, pos int, key, old, newer string) string

func AppendToXML(data map[string]string, output string, resolver CollisionResolver) (err error) {
	var lines []string
	lines, err = wkfs.ReadAllLines(output)
	if err != nil {
		return
	}

	newFileLines := make([]string, 0, len(lines))

	oldSet := make(map[string]*xmlLine)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<string name=") {
			m := &model.StringXMLItem{}
			err = xml.Unmarshal([]byte(trimmed), m)
			if err != nil {
				return
			}
			oldSet[m.Name] = &xmlLine{
				Key:   m.Name,
				Value: m.Value,
				Pos:   i,
			}
		} else if strings.HasPrefix(trimmed, "</resources>") {
			break
		}
		newFileLines = append(newFileLines, line)
	}

	newFileLines = append(newFileLines, "\n")

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
				if resolver != nil {
					value = resolver(output, oldEntry.Pos, key, oldEntry.Value, value)
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
			}
			continue
		}
		newFileLines = append(newFileLines, newLine)
	}

	newFileLines = append(newFileLines, "</resources>")

	sb := &strings.Builder{}
	for _, l := range newFileLines {
		sb.WriteString(l)
		sb.WriteRune('\n')
	}

	err = ioutil.WriteFile(output, []byte(sb.String()), 0644)

	return
}

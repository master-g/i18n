// Copyright © 2019 Master.G
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package i18n

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"github.com/master-g/i18n/pkg/wkio"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/master-g/i18n/internal/buildinfo"
	"github.com/sirupsen/logrus"
)

// Config of the csv converter
type Config struct {
	OutputJSONPath string
	Append         bool
	HasMeta        bool
	Overwrite      bool
}

// Metadata holds convention metadata
type Metadata struct {
	Version     string   `json:"tool_version"`
	SourceFiles []string `json:"source_files"`
	CreatedAt   string   `json:"created_at"`
	ModifiedAt  string   `json:"modified_at"`
}

// Converter holds context of this convention run
type Converter struct {
	cfg     *Config
	Meta    *Metadata                    `json:"metadata,omitempty"`
	Strings map[string]map[string]string `json:"strings"`
}

// NewConverter returns a new packer instance created with config
func NewConverter(cfg *Config) *Converter {
	if cfg == nil {
		return nil
	}

	return &Converter{
		cfg: cfg,
		Meta: &Metadata{
			Version:    buildinfo.VersionString(),
			CreatedAt:  time.Now().String(),
			ModifiedAt: time.Now().String(),
		},
		Strings: make(map[string]map[string]string),
	}
}

// ReadAppendFile to append old json file contents
func (converter *Converter) ReadAppendFile(path string) (err error) {
	var raw []byte
	raw, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	old := &Converter{}
	err = json.Unmarshal(raw, old)
	if err != nil {
		return
	}

	converter.Meta.SourceFiles = append(converter.Meta.SourceFiles, old.Meta.SourceFiles...)
	converter.Meta.CreatedAt = old.Meta.CreatedAt

	// merge strings
	for lang, key2str := range old.Strings {
		logrus.Debugf("old entry %v = %v", lang, key2str)
		converter.Strings[lang] = key2str
	}

	return nil
}

// Convert csv files
func (converter *Converter) Convert(csvFiles map[string]string) (err error) {
	// map to avoid source file duplication when append to old json file
	sourceFiles := make(map[string]string)
	for _, s := range converter.Meta.SourceFiles {
		sourceFiles[s] = s
	}

	// iterate source csv files
	for absPath, filename := range csvFiles {
		if _, ok := sourceFiles[filename]; !ok {
			sourceFiles[filename] = filename
			converter.Meta.SourceFiles = append(converter.Meta.SourceFiles, filename)
		}

		logrus.Debugf("processing %v", absPath)

		// read csv file
		var csvFile *os.File
		csvFile, err = os.Open(absPath)
		csvReader := csv.NewReader(bufio.NewReader(csvFile))

		// index to language
		index2language := make(map[int]string)

		for {
			var records []string
			records, err = csvReader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return
			}
			if len(index2language) == 0 {
				for i, lang := range records {
					if i > 0 && lang != "" {
						index2language[i] = lang
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
						strKey = str
					} else if lang, ok := index2language[i]; ok {
						if converter.Strings[lang] == nil {
							converter.Strings[lang] = make(map[string]string)
						}
						oldEntry, collision := converter.Strings[lang][strKey]
						if collision && oldEntry != records[i] {
							logrus.Warnf("lang %v has key %v collision in %v", lang, strKey, absPath)
						}
						converter.Strings[lang][strKey] = records[i]
					}
				}
			}
		}
	}
	if !converter.cfg.HasMeta {
		converter.Meta = nil
	}
	var resultJson []byte
	var outFile *os.File
	resultJson, err = json.Marshal(converter)
	if err != nil {
		return
	}
	outFile, err = os.Create(converter.cfg.OutputJSONPath)
	if err != nil {
		return
	}
	defer wkio.SafeClose(outFile)
	_, err = outFile.Write(resultJson)

	return
}

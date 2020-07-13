package model

import "encoding/json"

type LanguageKVS struct {
	Language string            `json:"language"`
	KVS      map[string]string `json:"kvs"`
}

type SourceFileType int

const (
	SourceFileTypeCSV = iota + 1
	// SourceFileTypeXML
	// SourceFileTypeXLS
)

type SourceFile struct {
	Type      SourceFileType          `json:"type"`
	AbsPath   string                  `json:"path"`
	Languages map[string]*LanguageKVS `json:"languages"`
}

func (s *SourceFile) String() string {
	raw, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	} else {
		return string(raw)
	}
}

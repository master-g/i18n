package model

import (
	"encoding/json"
)

type Collision struct {
	Language string   `json:"language"`
	Key      string   `json:"key"`
	Files    []string `json:"files"`
	Values   []string `json:"values"`
}

func (c *Collision) String() string {
	raw, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}

	return string(raw)
}

type MergeCollisionResolver func(collision *Collision) string

func Merge(sources []*SourceFile, resolver MergeCollisionResolver) map[string]map[string]string {
	temp := make(map[string]map[string]*Collision)

	for _, src := range sources {
		for lang, kvs := range src.Languages {
			if _, ok := temp[lang]; !ok {
				temp[lang] = make(map[string]*Collision)
			}

			for key, value := range kvs.KVS {
				if value == "" {
					continue
				}

				if _, ok := temp[lang][key]; !ok {
					temp[lang][key] = &Collision{
						Language: lang,
						Key:      key,
						Files:    nil,
						Values:   nil,
					}
				}

				c := temp[lang][key]

				if len(c.Values) == 0 {
					c.Files = append(c.Files, src.AbsPath)
					c.Values = append(c.Values, value)
				} else {
					found := false
					for _, v := range c.Values {
						if value == v {
							found = true
							break
						}
					}
					if !found {
						c.Files = append(c.Files, src.AbsPath)
						c.Values = append(c.Values, value)
					}
				}
			}
		}
	}

	result := make(map[string]map[string]string)
	for lang, collisions := range temp {
		result[lang] = make(map[string]string)

		for _, collision := range collisions {
			if len(collision.Values) == 0 {
				continue
			}

			var value string
			if len(collision.Values) > 1 && resolver != nil {
				value = resolver(collision)
			} else {
				value = collision.Values[0]
			}

			result[lang][collision.Key] = value
		}
	}

	return result
}

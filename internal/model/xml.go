package model

import "encoding/xml"

type StringXMLItem struct {
	XMLName xml.Name `xml:"string"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

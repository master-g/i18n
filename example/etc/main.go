package main

import (
	"encoding/xml"
	"fmt"
	"log"
)

type StringItem struct {
	XMLName xml.Name `xml:"string"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

func main() {
	s := &StringItem{
		Name:  "hello",
		Value: "world!",
	}

	raw, err := xml.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(raw))
}

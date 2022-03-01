package xml

import (
	"reflect"
	"strings"
)

type XMLTag struct {
	Name      string
	OmitEmpty bool
}

func parseXMLTag(field reflect.StructField) *XMLTag {
	xml_tag := &XMLTag{}
	if tag := field.Tag.Get("xml"); tag != "" {
		tokens := strings.Split(tag, ",")
		xml_tag.Name = tokens[0]
		// Only "omitempty" is currently supported; ignore unsupported flags
		for _, flag := range tokens[1:] {
			if flag == "omitempty" {
				xml_tag.OmitEmpty = true
				break
			}
		}
	}
	return xml_tag
}

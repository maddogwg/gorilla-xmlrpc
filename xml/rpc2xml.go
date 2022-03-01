// Copyright 2013 Ivan Danyliuk
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func rpcRequest2XML(method string, rpc interface{}) (string, error) {
	buffer := "<methodCall><methodName>"
	buffer += method
	buffer += "</methodName>"
	params, err := rpcParams2XML(rpc)
	buffer += params
	buffer += "</methodCall>"
	return buffer, err
}

func rpcResponse2XML(rpc interface{}) (string, error) {
	buffer := "<methodResponse>"
	params, err := rpcParams2XML(rpc)
	buffer += params
	buffer += "</methodResponse>"
	return buffer, err
}

func rpcParams2XML(rpc interface{}) (string, error) {
	var err error
	buffer := "<params>"
	for i := 0; i < reflect.ValueOf(rpc).Elem().NumField(); i++ {
		var xml string
		buffer += "<param>"
		xml, err = rpc2XML(reflect.ValueOf(rpc).Elem().Field(i).Interface(), false)
		buffer += xml
		buffer += "</param>"
	}
	buffer += "</params>"
	return buffer, err
}

func rpc2XML(value interface{}, omitEmpty bool) (string, error) {
	var out string

	switch reflect.ValueOf(value).Kind() {
	case reflect.Int:
		out = int2XML(value.(int), omitEmpty)
	case reflect.Float64:
		out = double2XML(value.(float64), omitEmpty)
	case reflect.String:
		out = string2XML(value.(string), omitEmpty)
	case reflect.Bool:
		out = bool2XML(value.(bool), omitEmpty)
	case reflect.Struct:
		if reflect.TypeOf(value).String() != "time.Time" {
			out = struct2XML(value, omitEmpty)
		} else {
			out = time2XML(value.(time.Time))
		}
	case reflect.Slice, reflect.Array:
		// FIXME: is it the best way to recognize '[]byte'?
		if reflect.TypeOf(value).String() != "[]uint8" {
			out = array2XML(value, omitEmpty)
		} else {
			out = base642XML(value.([]byte))
		}
	case reflect.Ptr:
		if reflect.ValueOf(value).IsNil() {
			if !omitEmpty {
				out = "<nil/>"
			}
		} else {
			// Omission only applies to indirect value when pointer is nil; no
			// need to propagate omitEmpty at this point.
			return rpc2XML(reflect.Indirect(reflect.ValueOf(value)).Interface(), false)
		}
	}

	if out != "" {
		return "<value>" + out + "</value>", nil
	}
	return "", nil
}

func int2XML(value int, omitEmpty bool) string {
	if omitEmpty && value == 0 {
		return ""
	}
	return fmt.Sprintf("<int>%d</int>", value)
}

func double2XML(value float64, omitEmpty bool) string {
	if omitEmpty && value == 0 {
		return ""
	}
	return fmt.Sprintf("<double>%f</double>", value)
}

func bool2XML(value bool, omitEmpty bool) string {
	if omitEmpty && !value {
		return ""
	}
	var b string
	if value {
		b = "1"
	} else {
		b = "0"
	}
	return fmt.Sprintf("<boolean>%s</boolean>", b)
}

func string2XML(value string, omitEmpty bool) string {
	if omitEmpty && value == "" {
		return ""
	}
	value = strings.Replace(value, "&", "&amp;", -1)
	value = strings.Replace(value, "\"", "&quot;", -1)
	value = strings.Replace(value, "<", "&lt;", -1)
	value = strings.Replace(value, ">", "&gt;", -1)
	return fmt.Sprintf("<string>%s</string>", value)
}

func struct2XML(value interface{}, omitEmpty bool) string {
	out := ""
	for i := 0; i < reflect.TypeOf(value).NumField(); i++ {
		field := reflect.ValueOf(value).Field(i)
		field_type := reflect.TypeOf(value).Field(i)
		var name string = field_type.Name
		field_tag := parseXMLTag(field_type)
		field_value, _ := rpc2XML(field.Interface(), field_tag.OmitEmpty)
		if field_value == "" {
			continue
		}
		if field_tag.Name != "" {
			name = field_tag.Name
		}
		field_name := fmt.Sprintf("<name>%s</name>", name)
		out += fmt.Sprintf("<member>%s%s</member>", field_name, field_value)
	}
	if out == "" {
		return ""
	}
	return "<struct>" + out + "</struct>"
}

func array2XML(value interface{}, omitEmpty bool) (out string) {
	if omitEmpty && reflect.ValueOf(value).Len() == 0 {
		out = ""
		return
	}
	out = "<array><data>"
	for i := 0; i < reflect.ValueOf(value).Len(); i++ {
		item_xml, _ := rpc2XML(reflect.ValueOf(value).Index(i).Interface(), false)
		out += item_xml
	}
	out += "</data></array>"
	return
}

func time2XML(t time.Time) string {
	/*
		// TODO: find out whether we need to deal
		// here with TZ
		var tz string;
		zone, offset := t.Zone()
		if zone == "UTC" {
			tz = "Z"
		} else {
			tz = fmt.Sprintf("%03d00", offset / 3600 )
		}
	*/
	return fmt.Sprintf("<dateTime.iso8601>%04d%02d%02dT%02d:%02d:%02d</dateTime.iso8601>",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

func base642XML(data []byte) string {
	str := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("<base64>%s</base64>", str)
}

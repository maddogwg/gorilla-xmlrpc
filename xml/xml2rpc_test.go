// Copyright 2013 Ivan Danyliuk
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type SubStructXml2Rpc struct {
	Foo  int
	Bar  string
	Data []int
}

type StructXml2Rpc struct {
	Int    int
	Float  float64
	Str    string
	Bool   bool
	Sub    SubStructXml2Rpc
	Time   time.Time
	Base64 []byte
}

func TestXML2RPC(t *testing.T) {
	req := new(StructXml2Rpc)
	err := xml2RPC("<methodCall><methodName>Some.Method</methodName><params><param><value><i4>123</i4></value></param><param><value><double>3.145926</double></value></param><param><value><string>Hello, World!</string></value></param><param><value><boolean>0</boolean></value></param><param><value><struct><member><name>Foo</name><value><int>42</int></value></member><member><name>Bar</name><value><string>I'm Bar</string></value></member><member><name>Data</name><value><array><data><value><int>1</int></value><value><int>2</int></value><value><int>3</int></value></data></array></value></member></struct></value></param><param><value><dateTime.iso8601>20120717T14:08:55</dateTime.iso8601></value></param><param><value><base64>eW91IGNhbid0IHJlYWQgdGhpcyE=</base64></value></param></params></methodCall>", req)
	if err != nil {
		t.Error("XML2RPC conversion failed", err)
	}
	expected_req := &StructXml2Rpc{123, 3.145926, "Hello, World!", false, SubStructXml2Rpc{42, "I'm Bar", []int{1, 2, 3}}, time.Date(2012, time.July, 17, 14, 8, 55, 0, time.Local), []byte("you can't read this!")}
	if !reflect.DeepEqual(req, expected_req) {
		t.Error("XML2RPC conversion failed")
		t.Error("Expected", expected_req)
		t.Error("Got", req)
	}
}

type StructSpecialCharsXml2Rpc struct {
	String1 string
}

func TestXML2RPCSpecialChars(t *testing.T) {
	req := new(StructSpecialCharsXml2Rpc)
	err := xml2RPC("<methodResponse><params><param><value><string> &amp; &quot; &lt; &gt; </string></value></param></params></methodResponse>", req)
	if err != nil {
		t.Error("XML2RPC conversion failed", err)
	}
	expected_req := &StructSpecialCharsXml2Rpc{" & \" < > "}
	if !reflect.DeepEqual(req, expected_req) {
		t.Error("XML2RPC conversion failed")
		t.Error("Expected", expected_req)
		t.Error("Got", req)
	}
}

type StructNilXml2Rpc struct {
	Ptr *int
}

func TestXML2RPCNil(t *testing.T) {
	req := new(StructNilXml2Rpc)
	err := xml2RPC("<methodResponse><params><param><value><nil/></value></param></params></methodResponse>", req)
	if err != nil {
		t.Error("XML2RPC conversion failed", err)
	}
	expected_req := &StructNilXml2Rpc{nil}
	if !reflect.DeepEqual(req, expected_req) {
		t.Error("XML2RPC conversion failed")
		t.Error("Expected", expected_req)
		t.Error("Got", req)
	}
}

type StructXml2RpcSubArgs struct {
	String1 string
	String2 string
	Id      int
}

type StructXml2RpcHelloArgs struct {
	Args StructXml2RpcSubArgs
}

func TestXML2RPCLowercasedMethods(t *testing.T) {
	req := new(StructXml2RpcHelloArgs)
	err := xml2RPC("<methodCall><params><param><value><struct><member><name>string1</name><value><string>I'm a first string</string></value></member><member><name>string2</name><value><string>I'm a second string</string></value></member><member><name>id</name><value><int>1</int></value></member></struct></value></param></params></methodCall>", req)
	if err != nil {
		t.Error("XML2RPC conversion failed", err)
	}
	args := StructXml2RpcSubArgs{"I'm a first string", "I'm a second string", 1}
	expected_req := &StructXml2RpcHelloArgs{args}
	if !reflect.DeepEqual(req, expected_req) {
		t.Error("XML2RPC conversion failed")
		t.Error("Expected", expected_req)
		t.Error("Got", req)
	}
}

func TestXML2PRCFaultCall(t *testing.T) {
	req := new(StructXml2RpcHelloArgs)
	data := `<?xmlversion="1.0"?><methodResponse><fault><value><struct><member><name>faultCode</name><value><int>116</int></value></member><member><name>faultString</name><value><string>Error
Requiredattribute'user'notfound:
[{'User',"gggg"},{'Host',"sss.com"},{'Password',"ssddfsdf"}]
</string></value></member></struct></value></fault></methodResponse>`

	errstr := `Error
Requiredattribute'user'notfound:
[{'User',"gggg"},{'Host',"sss.com"},{'Password',"ssddfsdf"}]
`

	err := xml2RPC(data, req)

	fault, ok := err.(Fault)
	if !ok {
		t.Errorf("error should be of concrete type Fault, but got %v", err)
	} else {
		if fault.Code != 116 {
			t.Errorf("expected fault.Code to be %d, but got %d", 116, fault.Code)
		}
		if fault.String != errstr {
			t.Errorf("fault.String should be:\n\n%s\n\nbut got:\n\n%s\n", errstr, fault.String)
		}
	}
}

func TestXML2PRCISO88591(t *testing.T) {
	req := new(StructXml2RpcHelloArgs)
	data := `<?xml version="1.0" encoding="ISO-8859-1"?><methodResponse><fault><value><struct><member><name>faultCode</name><value><int>116</int></value></member><member><name>faultString</name><value><string>Error
Requiredattribute'user'notfound:
[{'User',"` + "\xd6\xf1\xe4" + `"},{'Host',"sss.com"},{'Password',"ssddfsdf"}]
</string></value></member></struct></value></fault></methodResponse>`

	errstr := `Error
Requiredattribute'user'notfound:
[{'User',"Öñä"},{'Host',"sss.com"},{'Password',"ssddfsdf"}]
`

	err := xml2RPC(data, req)

	fault, ok := err.(Fault)
	if !ok {
		t.Errorf("error should be of concrete type Fault, but got %v", err)
	} else {
		if fault.Code != 116 {
			t.Errorf("expected fault.Code to be %d, but got %d", 116, fault.Code)
		}
		if fault.String != errstr {
			t.Errorf("fault.String should be:\n\n%s\n\nbut got:\n\n%s\n", errstr, fault.String)
		}
	}
}

type TaggedStructXml2RpcParams struct {
	Foo    string  `xml:""`                     // Empty tag
	Bar    int     `xml:"renameBar"`            // Rename, include if empty
	Str    string  `xml:",omitempty"`           // Use field name, omit if empty
	Double float64 `xml:"doublename,omitempty"` // Rename, omit if empty
	IntPtr *int    `xml:"ptrname,omitempty"`    // Rename, omit if empty
}

type TaggedStructXml2Rpc struct {
	Params TaggedStructXml2RpcParams
}

type TestXml2RpcTagsTest struct {
	Input  string
	Output *TaggedStructXml2Rpc
	err    error
}

func TestXML2RPCTags(t *testing.T) {
	var (
		intVal            int = 42
		smallestDoubleStr     = strconv.FormatFloat(math.SmallestNonzeroFloat64, 'E', -1, 64)
		tests                 = [...]TestXml2RpcTagsTest{
			{
				Input: "<methodCall>" +
					"<methodName>Test.Method</methodName>" +
					"<params><param><value><struct>" +
					"<member><name>Foo</name><value><string>FOO</string></value></member>" +
					"<member><name>renameBar</name><value><int>123</int></value></member>" +
					"<member><name>Str</name><value><string>STRING</string></value></member>" +
					"<member><name>doublename</name><value><double>" + smallestDoubleStr + "</double></value></member>" +
					"<member><name>ptrname</name><value><int>42</int></value></member>" +
					"</struct></value></param></params>" +
					"</methodCall>",
				Output: &TaggedStructXml2Rpc{
					Params: TaggedStructXml2RpcParams{
						Foo:    "FOO",
						Bar:    123,
						Str:    "STRING",
						Double: math.SmallestNonzeroFloat64,
						IntPtr: &intVal,
					},
				},
				err: nil,
			},
			{
				Input: "<methodCall>" +
					"<methodName>Test.Method</methodName>" +
					"<params><param><value><struct>" +
					"<member><name>Foo</name><value><string>FOO</string></value></member>" +
					"<member><name>renameBar</name><value><int>123</int></value></member>" +
					"</struct></value></param></params>" +
					"</methodCall>",
				Output: &TaggedStructXml2Rpc{
					Params: TaggedStructXml2RpcParams{
						Foo:    "FOO",
						Bar:    123,
						Str:    "",
						Double: 0,
						IntPtr: nil,
					},
				},
				err: nil,
			},
			{
				Input: "<methodCall>" +
					"<methodName>Test.Method</methodName>" +
					"<params><param><value><struct>" +
					"<member><name>Foo</name><value><string>FOO</string></value></member>" +
					"<member><name>nonextant</name><value><int>123</int></value></member>" +
					"<member><name>Str</name><value><string>STRING</string></value></member>" +
					"<member><name>doublename</name><value><double>" + smallestDoubleStr + "</double></value></member>" +
					"<member><name>ptrname</name><value><int>42</int></value></member>" +
					"</struct></value></param></params>" +
					"</methodCall>",
				Output: &TaggedStructXml2Rpc{
					Params: TaggedStructXml2RpcParams{
						Foo:    "FOO",
						Bar:    0, // Parsing should not stop before here
						Str:    "",
						Double: 0,
						IntPtr: nil,
					},
				},
				err: FaultApplicationError,
			},
		}
	)

	for i, test := range tests {
		req := new(TaggedStructXml2Rpc)
		err := xml2RPC(test.Input, req)
		if err != test.err {
			if test.err == nil {
				t.Errorf("XML2RPC Tagged structure conversion test %d failed: %v", i, err)
			} else {
				t.Errorf("XML2RPC Tagged structure conversion test %d did not trigger expected error", i)
				t.Error("Expected", test.err)
				t.Error("Got", err)
			}
		} else if !reflect.DeepEqual(req, test.Output) {
			t.Errorf("XML2RPC Tagged structure conversion test %d failed", i)
			t.Error("Expected", test.Output)
			t.Error("Got", req)
		}
	}
}

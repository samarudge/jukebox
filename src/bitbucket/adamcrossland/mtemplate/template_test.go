// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mtemplate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type Test struct {
	in, out, err string
}

type T struct {
	Item  string
	Value string
}

type U struct {
	Mp map[string]int
}

type S struct {
	Header        string
	HeaderPtr     *string
	Integer       int
	IntegerPtr    *int
	NilPtr        *int
	InnerT        T
	InnerPointerT *T
	Data          []T
	Pdata         []*T
	Empty         []*T
	Emptystring   string
	Null          []*T
	Vec           []string
	True          bool
	False         bool
	Mp            map[string]string
	JSON          interface{}
	Innermap      U
	Stringmap     map[string]string
	Ptrmap        map[string]*string
	Iface         interface{}
	Ifaceptr      interface{}
}

func (s *S) PointerMethod() string { return "ptrmethod!" }

func (s S) ValueMethod() string { return "valmethod!" }

var t1 = T{"ItemNumber1", "ValueNumber1"}
var t2 = T{"ItemNumber2", "ValueNumber2"}

func uppercase(v interface{}) string {
	s := v.(string)
	t := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'a' <= c && c <= 'z' {
			c = c + 'A' - 'a'
		}
		t += string(c)
	}
	return t
}

func plus1(v interface{}) string {
	i := v.(int)
	return fmt.Sprint(i + 1)
}

func writer(f func(interface{}) string) func(io.Writer, string, ...interface{}) {
	return func(w io.Writer, format string, v ...interface{}) {
		if len(v) != 1 {
			panic("test writer expected one arg")
		}
		io.WriteString(w, f(v[0]))
	}
}

func multiword(w io.Writer, format string, value ...interface{}) {
	for _, v := range value {
		fmt.Fprintf(w, "<%v>", v)
	}
}

var formatters = FormatterMap{
	"uppercase": writer(uppercase),
	"+1":        writer(plus1),
	"multiword": multiword,
}

var tests = []*Test{
	// Simple
	&Test{"", "", ""},
	&Test{"abc", "abc", ""},
	&Test{"abc\ndef\n", "abc\ndef\n", ""},
	&Test{" {.meta-left}   \n", "{", ""},
	&Test{" {.meta-right}   \n", "}", ""},
	&Test{" {.space}   \n", " ", ""},
	&Test{" {.tab}   \n", "\t", ""},
	&Test{"     {#comment}   \n", "", ""},
	&Test{"\tSome Text\t\n", "\tSome Text\t\n", ""},
	&Test{" {.meta-right} {.meta-right} {.meta-right} \n", " } } } \n", ""},

	// Variables at top level
	&Test{
		in: "{!Header}={!Integer}\n",

		out: "Header=77\n",
	},

	&Test{
		in: "Pointers: {!*HeaderPtr}={!*IntegerPtr}\n",

		out: "Pointers: Header=77\n",
	},

	&Test{
		in: "Stars but not pointers: {!*Header}={!*Integer}\n",

		out: "Stars but not pointers: Header=77\n",
	},

	&Test{
		in: "nil pointer: {!*NilPtr}={!*Integer}\n",

		out: "nil pointer: <nil>=77\n",
	},

	// Method at top level
	&Test{
		in: "ptrmethod={!PointerMethod}\n",

		out: "ptrmethod=ptrmethod!\n",
	},

	&Test{
		in: "valmethod={!ValueMethod}\n",

		out: "valmethod=valmethod!\n",
	},

	// Section
	&Test{
		in: "{.section Data }\n" +
			"some text for the section\n" +
			"{.end}\n",

		out: "some text for the section\n",
	},
	&Test{
		in: "{.section Data }\n" +
			"{!Header}={!Integer}\n" +
			"{.end}\n",

		out: "Header=77\n",
	},
	&Test{
		in: "{.section Pdata }\n" +
			"{!Header}={!Integer}\n" +
			"{.end}\n",

		out: "Header=77\n",
	},
	&Test{
		in: "{.section Pdata }\n" +
			"data present\n" +
			"{.or}\n" +
			"data not present\n" +
			"{.end}\n",

		out: "data present\n",
	},
	&Test{
		in: "{.section Empty }\n" +
			"data present\n" +
			"{.or}\n" +
			"data not present\n" +
			"{.end}\n",

		out: "data not present\n",
	},
	&Test{
		in: "{.section Null }\n" +
			"data present\n" +
			"{.or}\n" +
			"data not present\n" +
			"{.end}\n",

		out: "data not present\n",
	},
	&Test{
		in: "{.section Pdata }\n" +
			"{!Header}={!Integer}\n" +
			"{.section @ }\n" +
			"{!Header}={!Integer}\n" +
			"{.end}\n" +
			"{.end}\n",

		out: "Header=77\n" +
			"Header=77\n",
	},

	&Test{
		in: "{.section Data}{.end} {!Header}\n",

		out: " Header\n",
	},

	&Test{
		in: "{.section Integer}{!@}{.end}",

		out: "77",
	},


	// Repeated
	&Test{
		in: "{.section Pdata }\n" +
			"{.repeated section @ }\n" +
			"{!Item}={!Value}\n" +
			"{.end}\n" +
			"{.end}\n",

		out: "ItemNumber1=ValueNumber1\n" +
			"ItemNumber2=ValueNumber2\n",
	},
	&Test{
		in: "{.section Pdata }\n" +
			"{.repeated section @ }\n" +
			"{!Item}={!Value}\n" +
			"{.or}\n" +
			"this should not appear\n" +
			"{.end}\n" +
			"{.end}\n",

		out: "ItemNumber1=ValueNumber1\n" +
			"ItemNumber2=ValueNumber2\n",
	},
	&Test{
		in: "{.section @ }\n" +
			"{.repeated section Empty }\n" +
			"{!Item}={!Value}\n" +
			"{.or}\n" +
			"this should appear: empty field\n" +
			"{.end}\n" +
			"{.end}\n",

		out: "this should appear: empty field\n",
	},
	&Test{
		in: "{.repeated section Pdata }\n" +
			"{!Item}\n" +
			"{.alternates with}\n" +
			"is\nover\nmultiple\nlines\n" +
			"{.end}\n",

		out: "ItemNumber1\n" +
			"is\nover\nmultiple\nlines\n" +
			"ItemNumber2\n",
	},
	&Test{
		in: "{.repeated section Pdata }\n" +
			"{!Item}\n" +
			"{.alternates with}\n" +
			"is\nover\nmultiple\nlines\n" +
			" {.end}\n",

		out: "ItemNumber1\n" +
			"is\nover\nmultiple\nlines\n" +
			"ItemNumber2\n",
	},
	&Test{
		in: "{.section Pdata }\n" +
			"{.repeated section @ }\n" +
			"{!Item}={!Value}\n" +
			"{.alternates with}DIVIDER\n" +
			"{.or}\n" +
			"this should not appear\n" +
			"{.end}\n" +
			"{.end}\n",

		out: "ItemNumber1=ValueNumber1\n" +
			"DIVIDER\n" +
			"ItemNumber2=ValueNumber2\n",
	},
	&Test{
		in: "{.repeated section Vec }\n" +
			"{!@}\n" +
			"{.end}\n",

		out: "elt1\n" +
			"elt2\n",
	},
	// Same but with a space before {.end}: was a bug.
	&Test{
		in: "{.repeated section Vec }\n" +
			"{!@} {.end}\n",

		out: "elt1 elt2 \n",
	},
	&Test{
		in: "{.repeated section Integer}{.end}",

		err: "line 1: .repeated: cannot repeat Integer (type int)",
	},

	// Nested names
	&Test{
		in: "{.section @ }\n" +
			"{!InnerT.Item}={!InnerT.Value}\n" +
			"{.end}",

		out: "ItemNumber1=ValueNumber1\n",
	},
	&Test{
		in: "{.section @ }\n" +
			"{!InnerT.Item}={.section InnerT}{.section Value}{!@}{.end}{.end}\n" +
			"{.end}",

		out: "ItemNumber1=ValueNumber1\n",
	},

	&Test{
		in: "{.section Emptystring}emptystring{.end}\n" +
			"{.section Header}header{.end}\n",

		out: "\nheader\n",
	},

	&Test{
		in: "{.section True}1{.or}2{.end}\n" +
			"{.section False}3{.or}4{.end}\n",

		out: "1\n4\n",
	},

	// Maps

	&Test{
		in: "{!Mp.mapkey}\n",

		out: "Ahoy!\n",
	},
	&Test{
		in: "{!Innermap.Mp.innerkey}\n",

		out: "55\n",
	},
	&Test{
		in: "{.section Innermap}{.section Mp}{!innerkey}{.end}{.end}\n",

		out: "55\n",
	},
	&Test{
		in: "{.section JSON}{.repeated section maps}{!a}{!b}{.end}{.end}\n",

		out: "1234\n",
	},
	&Test{
		in: "{!Stringmap.stringkey1}\n",

		out: "stringresult\n",
	},
	&Test{
		in: "{.repeated section Stringmap}\n" +
			"{!@}\n" +
			"{.end}",

		out: "stringresult\n" +
			"stringresult\n",
	},
	&Test{
		in: "{.repeated section Stringmap}\n" +
			"\t{!@}\n" +
			"{.end}",

		out: "\tstringresult\n" +
			"\tstringresult\n",
	},
	&Test{
		in: "{!*Ptrmap.stringkey1}\n",

		out: "pointedToString\n",
	},
	&Test{
		in: "{.repeated section Ptrmap}\n" +
			"{!*@}\n" +
			"{.end}",

		out: "pointedToString\n" +
			"pointedToString\n",
	},


	// Interface values

	&Test{
		in: "{!Iface}",

		out: "[1 2 3]",
	},
	&Test{
		in: "{.repeated section Iface}{!@}{.alternates with} {.end}",

		out: "1 2 3",
	},
	&Test{
		in: "{.section Iface}{!@}{.end}",

		out: "[1 2 3]",
	},
	&Test{
		in: "{.section Ifaceptr}{!Item} {!Value}{.end}",

		out: "Item Value",
	},
}

func TestAll(t *testing.T) {
	// Parse
	testAll(t, func(test *Test) (*Template, error) { return Parse(test.in, formatters) })
	// ParseFile
	testAll(t, func(test *Test) (*Template, error) {
		err := ioutil.WriteFile("_test/test.tmpl", []byte(test.in), 0600)
		if err != nil {
			t.Error("unexpected write error:", err)
			return nil, err
		}
        // Clear out the cache in between eahc parse. Otherwise, the
        // tests will all break.
        delete(parsedCache, "_test/test.tmpl")
		return ParseFile("_test/test.tmpl", formatters)
	})
	// tmpl.ParseFile
	testAll(t, func(test *Test) (*Template, error) {
		err := ioutil.WriteFile("_test/test.tmpl", []byte(test.in), 0600)
		if err != nil {
			t.Error("unexpected write error:", err.Error())
			return nil, err
		}
		tmpl := New(formatters)
		return tmpl, tmpl.ParseFile("_test/test.tmpl")
	})
    // One more time, to make sure that we have an empty cache.
    delete(parsedCache, "_test/test.tmpl")
}

func testAll(t *testing.T, parseFunc func(*Test) (*Template, error)) {
	s := new(S)
	// initialized by hand for clarity.
	s.Header = "Header"
	s.HeaderPtr = &s.Header
	s.Integer = 77
	s.IntegerPtr = &s.Integer
	s.InnerT = t1
	s.Data = []T{t1, t2}
	s.Pdata = []*T{&t1, &t2}
	s.Empty = []*T{}
	s.Null = nil
	s.Vec = make([]string, 2)
	s.Vec[0] = "elt1"
	s.Vec[1] = "elt2"
	s.True = true
	s.False = false
	s.Mp = make(map[string]string)
	s.Mp["mapkey"] = "Ahoy!"
	json.Unmarshal([]byte(`{"maps":[{"a":1,"b":2},{"a":3,"b":4}]}`), &s.JSON)
	s.Innermap.Mp = make(map[string]int)
	s.Innermap.Mp["innerkey"] = 55
	s.Stringmap = make(map[string]string)
	s.Stringmap["stringkey1"] = "stringresult" // the same value so repeated section is order-independent
	s.Stringmap["stringkey2"] = "stringresult"
	s.Ptrmap = make(map[string]*string)
	x := "pointedToString"
	s.Ptrmap["stringkey1"] = &x // the same value so repeated section is order-independent
	s.Ptrmap["stringkey2"] = &x
	s.Iface = []int{1, 2, 3}
	s.Ifaceptr = &T{"Item", "Value"}

	var buf bytes.Buffer
	for _, test := range tests {
		buf.Reset()
		tmpl, err := parseFunc(test)
		if err != nil {
			t.Error("unexpected parse error: ", err)
			continue
		}
		err = tmpl.Execute(&buf, s)
		if test.err == "" {
			if err != nil {
				//t.Error("unexpected execute error:", err)
				fmt.Printf("unexpected execute error:", err)
			}
		} else {
			if err == nil {
				t.Errorf("expected execute error %q, got nil", test.err)
				//fmt.Printf("expected execute error %q, got nil", test.err)
			} else if err.Error() != test.err {
				t.Errorf("expected execute error %q, got %q", test.err, err.Error())
				//fmt.Printf("expected execute error %q, got %q", test.err, err.String())
			}
		}
		if buf.String() != test.out {
			t.Errorf("for %q: expected %q got %q", test.in, test.out, buf.String())
			fmt.Printf("for %q: expected %q got %q", test.in, test.out, buf.String())
		}
	}
}

func TestMapDriverType(t *testing.T) {
	mp := map[string]string{"footer": "Ahoy!"}
	tmpl, err := Parse("template: {!footer}", nil)
	if err != nil {
		t.Error("unexpected parse error:", err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, mp)
	if err != nil {
		t.Error("unexpected execute error:", err)
	}
	s := b.String()
	expect := "template: Ahoy!"
	if s != expect {
		t.Errorf("failed passing string as data: expected %q got %q", expect, s)
	}
}

func TestMapNoEntry(t *testing.T) {
	mp := make(map[string]int)
	tmpl, err := Parse("template: {!notthere}!", nil)
	if err != nil {
		t.Error("unexpected parse error:", err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, mp)
	if err != nil {
		t.Error("unexpected execute error:", err)
	}
	s := b.String()
	expect := "template: 0!"
	if s != expect {
		t.Errorf("failed passing string as data: expected %q got %q", expect, s)
	}
}

func TestStringDriverType(t *testing.T) {
	tmpl, err := Parse("template: {!@}", nil)
	if err != nil {
		t.Error("unexpected parse error:", err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, "hello")
	if err != nil {
		t.Error("unexpected execute error:", err)
	}
	s := b.String()
	expect := "template: hello"
	if s != expect {
		t.Errorf("failed passing string as data: expected %q got %q", expect, s)
	}
}

func TestTwice(t *testing.T) {
	tmpl, err := Parse("template: {!@}", nil)
	if err != nil {
		t.Error("unexpected parse error:", err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, "hello")
	if err != nil {
		t.Error("unexpected parse error:", err)
	}
	s := b.String()
	expect := "template: hello"
	if s != expect {
		t.Errorf("failed passing string as data: expected %q got %q", expect, s)
	}
	err = tmpl.Execute(&b, "hello")
	if err != nil {
		t.Error("unexpected parse error:", err)
	}
	s = b.String()
	expect += expect
	if s != expect {
		t.Errorf("failed passing string as data: expected %q got %q", expect, s)
	}
}

// Test that a variable evaluates to the field itself and does not further indirection
func TestVarIndirection(t *testing.T) {
	s := new(S)
	// initialized by hand for clarity.
	s.InnerPointerT = &t1

	var buf bytes.Buffer
	input := "{.section @}{!InnerPointerT}{.end}"
	tmpl, err := Parse(input, nil)
	if err != nil {
		t.Fatal("unexpected parse error:", err)
	}
	err = tmpl.Execute(&buf, s)
	if err != nil {
		t.Fatal("unexpected execute error:", err)
	}
	expect := fmt.Sprintf("%v", &t1) // output should be hex address of t1
	if buf.String() != expect {
		t.Errorf("for %q: expected %q got %q", input, expect, buf.String())
	}
}

func TestHTMLFormatterWithByte(t *testing.T) {
	s := "Test string."
	b := []byte(s)
	var buf bytes.Buffer
	HTMLFormatter(&buf, "", b)
	bs := buf.String()
	if bs != s {
		t.Errorf("munged []byte, expected: %s got: %s", s, bs)
	}
}

type UF struct {
	I int
	s string
}

func TestReferenceToUnexported(t *testing.T) {
	u := &UF{3, "hello"}
	var buf bytes.Buffer
	input := "{.section @}{!I}{!s}{.end}"
	tmpl, err := Parse(input, nil)
	if err != nil {
		t.Fatal("unexpected parse error:", err)
	}
	err = tmpl.Execute(&buf, u)
	if err == nil {
		t.Fatal("expected execute error, got none")
	}
	if strings.Index(err.Error(), "not exported") < 0 {
		t.Fatal("expected unexported error; got", err)
	}
}

var formatterTests = []Test{
	{
		in: "{!Header|uppercase}={!Integer|+1}\n" +
			"{!Header|html}={!Integer|str}\n",

		out: "HEADER=78\n" +
			"Header=77\n",
	},

	{
		in: "{!Header|uppercase}={!Integer Header|multiword}\n" +
			"{!Header|html}={!Header Integer|multiword}\n" +
			"{!Header|html}={!Header Integer}\n",

		out: "HEADER=<77><Header>\n" +
			"Header=<Header><77>\n" +
			"Header=Header77\n",
	},
	{
		in: "{!Raw}\n" +
			"{!Raw|html}\n",

		out: "a <&> b\n" +
			"a &lt;&amp;&gt; b\n",
	},
	{
		in:  "{!Bytes}",
		out: "hello",
	},
	{
		in:  "{!Raw|uppercase|html|html}",
		out: "A &amp;lt;&amp;amp;&amp;gt; B",
	},
	{
		in:  "{!Header Integer|multiword|html}",
		out: "&lt;Header&gt;&lt;77&gt;",
	},
	{
		in:  "{!Integer|no_formatter|html}",
		err: `unknown formatter: "no_formatter"`,
	},
	{
		in:  "{!Integer|||||}", // empty string is a valid formatter
		out: "77",
	},
}

func TestFormatters(t *testing.T) {
	data := map[string]interface{}{
		"Header":  "Header",
		"Integer": 77,
		"Raw":     "a <&> b",
		"Bytes":   []byte("hello"),
	}
	for _, c := range formatterTests {
		tmpl, err := Parse(c.in, formatters)
		if err != nil {
			if c.err == "" {
				t.Error("unexpected parse error:", err)
				continue
			}
			if strings.Index(err.Error(), c.err) < 0 {
				t.Errorf("unexpected error: expected %q, got %q", c.err, err.Error())
				continue
			}
		} else {
			if c.err != "" {
				t.Errorf("For %q, expected error, got none.", c.in)
				continue
			}
			buf := bytes.NewBuffer(nil)
			err = tmpl.Execute(buf, data)
			if err != nil {
				t.Error("unexpected Execute error: ", err)
				continue
			}
			actual := buf.String()
			if actual != c.out {
				t.Errorf("for %q: expected %q but got %q.", c.in, c.out, actual)
			}
		}
	}
}

func TestParentPageSimple(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}I am the Child Page"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("Parent Page Begin - {.child} - Parent Page End"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")
    
	testTemplate, _ := ParseFile("_test/child.html", nil)
	readBuf := new(bytes.Buffer)
	testTemplate.Execute(readBuf, nil)
	
	expected := "Parent Page Begin - I am the Child Page - Parent Page End"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestParentPageVariables(t *testing.T) {
	testTemplate, _ := ParseWithParentPage("{.parent}{!body}", "{!title} - {.child} - {!footer}", nil)
	readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
	testTemplate.Execute(readBuf, vars)
	
    
	expected := "Parent Page Begin - I am the Child Page - Parent Page End"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestParentPageVariablesFile(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}{!body}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{!title} - {.child} - {!footer}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")
    
	testTemplate, _ := ParseFile("_test/child.html", nil)
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
	testTemplate.Execute(readBuf, vars)
	
    
	expected := "Parent Page Begin - I am the Child Page - Parent Page End"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestMultiParent(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}<h3>Child template.</h3>"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{.parent _test/parent_parent.html}<h2>Parent Page</h2>{.child}<h2>/Parent Page</h2>"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")

	err = ioutil.WriteFile("_test/parent_parent.html", []byte("<h1>Outer Parent Page</h1>{.child}<h1>/Outer Parent Page</h1>"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent_parent.html: ", err)
    }
    defer os.Remove("_test/parent_parent.html")
    defer ClearFromCache("_test/parent_parent.html")
    
	testTemplate, _ := ParseFile("_test/child.html", nil)
    readBuf := new(bytes.Buffer)
	testTemplate.Execute(readBuf, nil)
	
    
	expected := "<h1>Outer Parent Page</h1><h2>Parent Page</h2><h3>Child template.</h3><h2>/Parent Page</h2><h1>/Outer Parent Page</h1>"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestCachingInit(t *testing.T) {
    if len(parsedCache) != 0 {
        t.Errorf("parsedCache has a length of %d, but it should be 0", len(parsedCache))
    }
}

func TestCachingFiles(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}{!body}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{!title} - {.child} - {!footer}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    
	testTemplate, _ := ParseFile("_test/child.html", nil)
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
	testTemplate.Execute(readBuf, vars)
	defer ClearFromCache("_test/child.html")
    defer ClearFromCache("_test/parent.html")

	// Make sure that the result of the templates being executed is correct
	expected := "Parent Page Begin - I am the Child Page - Parent Page End"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
	
	// The cache should contain two files
    if len(parsedCache) != 2 {
        t.Errorf("parsedCache has a length of %d, but it should be 2", len(parsedCache))
    }

	// Since the template are cached, we should be able to delete them
	// from the disk and still be able to execute the files.
    os.Remove("_test/child.html")
    os.Remove("_test/parent.html")
	retestTemplate, _ := ParseFile("_test/child.html", nil)
    rereadBuf := new(bytes.Buffer)
    vars = map[string]string{"title": "Parent Page Re-begin", "footer": "Parent Page Re-end", "body": "I am the re-child Page",}
	retestTemplate.Execute(rereadBuf, vars)

	// Make sure that the result of the templates being re-executed is correct
	reexpected := "Parent Page Re-begin - I am the re-child Page - Parent Page Re-end"
	if rereadBuf.String() != reexpected {
		t.Errorf("Expected to get '%s'\n but got '%s' from the second execution of the test.", reexpected, rereadBuf.String())
	}
}

func TestRender(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}{!body}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{!title} - {.child} - {!footer}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")
    
    
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
    RenderFile("_test/child.html", readBuf, vars)
    
	expected := "Parent Page Begin - I am the Child Page - Parent Page End"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestRenderWithBlocks(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}{.block foo}Foo:{!body}{.end}{.block footer}//This is the end.{.end}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{!title} - {.child foo} - {!footer}{.child footer}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")
    
    
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
    RenderFile("_test/child.html", readBuf, vars)
    
	expected := "Parent Page Begin - Foo:I am the Child Page - Parent Page End//This is the end."
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestRenderWithBlocksMultiLevel(t *testing.T) {
    err := ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}{.block foo}Foo:{!body}{.end}{.block footer}//This is the end.{.end}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{.parent _test/parent_parent.html}{!title} - {.child foo} - {!footer}{.child footer}{.block title}Test Everything{.end}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")
    
	err = ioutil.WriteFile("_test/parent_parent.html", []byte("<html><head><title>{.child title}</title></head><body>{.child}</body></html>"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent_parent.html")
    defer ClearFromCache("_test/parent_parent.html")
    
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
    RenderFile("_test/child.html", readBuf, vars)
    
	expected := "<html><head><title>Test Everything</title></head><body>Parent Page Begin - Foo:I am the Child Page - Parent Page End//This is the end.</body></html>"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestInclude(t *testing.T) {
    err := ioutil.WriteFile("_test/footer.html", []byte("<hr/>This is a static footer."), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/footer.html: ", err)
    }
    defer os.Remove("_test/footer.html")
    defer ClearFromCache("_test/footer.html")
    
	err = ioutil.WriteFile("_test/test.html", []byte("<html><body>This is a body of static text.{.include _test/footer.html}</body></html>"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/test.html: ", err)
    }
    defer os.Remove("_test/test.html")
    defer ClearFromCache("_test/test.html")
    
    
    readBuf := new(bytes.Buffer)
    RenderFile("_test/test.html", readBuf, nil)
    
	expected := "<html><body>This is a body of static text.<hr/>This is a static footer.</body></html>"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestIncludeWithBlocks(t *testing.T) {
    err := ioutil.WriteFile("_test/footer.html", []byte("<hr/>This is a static footer."), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/footer.html: ", err)
    }
    defer os.Remove("_test/footer.html")
    defer ClearFromCache("_test/footer.html")
    
    err = ioutil.WriteFile("_test/child.html", []byte("{.parent _test/parent.html}{.block foo}Foo:{!body}{.end}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/child.html")
    defer ClearFromCache("_test/child.html")
    
	err = ioutil.WriteFile("_test/parent.html", []byte("{!title} - {.child foo} - {!footer}{.include _test/footer.html}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/parent.html: ", err)
    }
    defer os.Remove("_test/parent.html")
    defer ClearFromCache("_test/parent.html")
    
    
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"title": "Parent Page Begin", "footer": "Parent Page End", "body": "I am the Child Page",}
    RenderFile("_test/child.html", readBuf, vars)
    
	expected := "Parent Page Begin - Foo:I am the Child Page - Parent Page End<hr/>This is a static footer."
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestJavaScript(t *testing.T) {
    err := ioutil.WriteFile("_test/js.html", []byte("<script>\n\tfunction foo() {\n\t\talert('{!message}');\n\t}\n</script>{!body}"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/js.html: ", err)
    }
    defer os.Remove("_test/js.html")
    defer ClearFromCache("_test/js.html")    
    
    readBuf := new(bytes.Buffer)
    vars := map[string]string{"message": "I am an alert message.", "body": "Hello, world!"}
    RenderFile("_test/js.html", readBuf, vars)
    
	expected := "<script>\n\tfunction foo() {\n\t\talert('I am an alert message.');\n\t}\n</script>Hello, world!"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
}

func TestRenderWithIntegers(t *testing.T) {
    err := ioutil.WriteFile("_test/ints.html", []byte("var foo=[a: {!a}, b: {! b}];"), 0600)
    if err != nil {
        t.Error("Unexpected error writing file _test/child.html: ", err)
    }
    defer os.Remove("_test/ints.html")
    defer ClearFromCache("_test/ints.html")    
    
    readBuf := new(bytes.Buffer)
    vars := map[string]int{"a": 40, "b": 20}
    RenderFile("_test/ints.html", readBuf, vars)
    
	expected := "var foo=[a: 40, b: 20];"
	if readBuf.String() != expected {
		t.Errorf("Expected to get '%s'\n but got '%s'", expected, readBuf.String())
	}
	
}
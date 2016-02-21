// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modifications by Adam Crossland. The Go Authors might not want to
// have anything to do with the changes that I have made within.
// They would almost certainly prefer that you use the exp/template
// package, which provides sophisticated template operations in a
// different way. Go ahead and check it out; I won't take it personally.
// I have attempted to note where I have made changes, but you can always
// diff against $(GOROOT)/src/pkg/template/template.go

/*
	Package template implements data-driven templates for generating textual
	output such as HTML.

	Templates are executed by applying them to a data structure.
	Annotations in the template refer to elements of the data
	structure (typically a field of a struct or a key in a map)
	to control execution and derive values to be displayed.
	The template walks the structure as it executes and the
	"cursor" @ represents the value at the current location
	in the structure.

	Data items may be values or pointers; the interface hides the
	indirection.

	In the following, 'field' is one of several things, according to the data.

		- The name of a field of a struct (result = data.field),
		- The value stored in a map under that key (result = data[field]), or
		- The result of invoking a niladic single-valued method with that name
		  (result = data.field())

	Major constructs ({} are the default delimiters for template actions;
	[] are the notation in this comment for optional elements):

		{# comment }

	A one-line comment.

		{.section field} XXX [ {.or} YYY ] {.end}

	Set @ to the value of the field.  It may be an explicit @
	to stay at the same point in the data. If the field is nil
	or empty, execute YYY; otherwise execute XXX.

		{.repeated section field} XXX [ {.alternates with} ZZZ ] [ {.or} YYY ] {.end}

	Like .section, but field must be an array or slice.  XXX
	is executed for each element.  If the array is nil or empty,
	YYY is executed instead.  If the {.alternates with} marker
	is present, ZZZ is executed between iterations of XXX.

		{! field}
		{! field1 field2 ...}
		{! field|formatter}
		{! field1 field2...|formatter}
		{! field|formatter1|formatter2}

	Insert the value of the fields into the output. Each field is
	first looked for in the cursor, as in .section and .repeated.
	If it is not found, the search continues in outer sections
	until the top level is reached.

	If the field value is a pointer, leading asterisks indicate
	that the value to be inserted should be evaluated through the
	pointer.  For example, if x.p is of type *int, {x.p} will
	insert the value of the pointer but {*x.p} will insert the
	value of the underlying integer.  If the value is nil or not a
	pointer, asterisks have no effect.

	If a formatter is specified, it must be named in the formatter
	map passed to the template set up routines or in the default
	set ("html","str","") and is used to process the data for
	output.  The formatter function has signature
		func(wr io.Writer, formatter string, data ...interface{})
	where wr is the destination for output, data holds the field
	values at the instantiation, and formatter is its name at
	the invocation site.  The default formatter just concatenates
	the string representations of the fields.

	Multiple formatters separated by the pipeline character | are
	executed sequentially, with each formatter receiving the bytes
	emitted by the one to its left.

	The delimiter strings get their default value, "{" and "}", from
	JSON-template.  They may be set to any non-empty, space-free
	string using the SetDelims method.  Their value can be printed
	in the output using {.meta-left} and {.meta-right}.

    mtemplate:
    Django-style template inheritance can be achieved by using the
    .parent and .child directives.

    A template that expects to be rendered inside of another template
    declares which template with the .parent directive, and a template
    that will have other templates rendered inside of it uses the 
    .child directive to indicate where the included template appears.

    For example, if I have a template called detail.html:
        {.parent masterpage.html}
        <h3>Child template.</h3>

    and a template called masterpage.html:
        <h2>Parent Page</h2>
        {.child}
        <h2>/Parent Page</h2>

    rendering detail.html like this:
        theTemplate := mtemplate.MustParseFile("detail.html", nil)
        outBuffer := new(bytes.Buffer)
        theTemplate.Execute(outBuffer, nil)

    will fill outBuffer with:
        <h2>Parent Page</h2>
        <h3>Child template.</h3>        
        <h2>/Parent Page</h2>

    templates can be nested in this way as deeply as the developer
    needs.

*/
package mtemplate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Errors returned during parsing and execution.  Users may extract the information and reformat
// if they desire.
type Error struct {
	Line int
	Msg  string
}

func (e *Error) Error() string { return fmt.Sprintf("line %d: %s", e.Line, e.Msg) }

// Most of the literals are aces.
var lbrace = []byte{'{'}
var rbrace = []byte{'}'}
var space = []byte{' '}
var tab = []byte{'\t'}

// The various types of "tokens", which are plain text or (usually) brace-delimited descriptors
const (
	tokAlternates = iota
	tokComment
	tokEnd
	tokLiteral
	tokOr
	tokRepeated
	tokSection
	tokText
	tokVariable
	tokParent // added for mtemplate
	tokChild
	tokBlock
	tokInclude
)

// FormatterMap is the type describing the mapping from formatter
// names to the functions that implement them.
type FormatterMap map[string]func(io.Writer, string, ...interface{})

// Built-in formatters.
var builtins = FormatterMap{
	"html":    HTMLFormatter,
	"str":     StringFormatter,
	"urlsafe": UrlFormatter,
	"":        StringFormatter,
}

// The parsed state of a template is a vector of xxxElement structs.
// Sections have line numbers so errors can be reported better during execution.

// Plain text.
type textElement struct {
	text []byte
}

// A literal such as .meta-left or .meta-right
type literalElement struct {
	text []byte
}

// A variable invocation to be evaluated
type variableElement struct {
	linenum int
	word    []string // The fields in the invocation.
	fmts    []string // Names of formatters to apply. len(fmts) > 0
}

// A .section block, possibly with a .or
type sectionElement struct {
	linenum int    // of .section itself
	field   string // cursor field for this block
	start   int    // first element
	or      int    // first element of .or block
	end     int    // one beyond last element
}

// A .repeated block, possibly with a .or and a .alternates
type repeatedElement struct {
	sectionElement     // It has the same structure...
	altstart       int // ... except for alternates
	altend         int
}

// mtemplate: A .parent block. It may hold either the name of the file
// in which the parent page template resides, or it may hold the
// template itself. 
type parentElement struct {
	filename string // The file in which the parent template is found
	body     string // Alternatively, the parent can be provided
	// as a string rather than as the name of a file
}

// mtemplate: A .child block. This type is used for parsing only.
type childElement struct {
	name string // The optional name of the child block to insert
}

type blockElement struct {
	name  string         // The name of the block
	elems *elemlist // The elements that belong to this block
}

type includeElement struct {
	fileName string         // The name of the file that is to be included
	elems    *elemlist // The elements that are to be included
}

// Template is the type that represents a template definition.
// It is unchanged after parsing.
type Template struct {
	fmap FormatterMap // formatters for variables
	// Used during parsing:
	ldelim, rdelim []byte // delimiters; default {}
	buf            []byte // input text to process
	p              int    // position in buf
	linenum        int    // position in input
	// Parsed results:
	elems *elemlist
	// mtemplate:
	// Used during execution
	parent    *parentElement
	childData map[string]*bytes.Buffer
	blockData map[string]*bytes.Buffer
}

// mtemplate: in order to make Templates thread-safe, we need to make
// a copy for execution, as parent
func (t Template) clone() (copyT *Template) {
	copyT = new(Template)
	copyT.elems = t.elems
	copyT.parent = t.parent
	copyT.childData = t.childData
	copyT.blockData = t.blockData

	return copyT
}

// Internal state for executing a Template.  As we evaluate the struct,
// the data item descends into the fields associated with sections, etc.
// Parent is used to walk upwards to find variables higher in the tree.
type state struct {
	parent *state          // parent in hierarchy
	data   reflect.Value   // the driver data for this section etc.
	wr     io.Writer       // where to send output
	buf    [2]bytes.Buffer // alternating buffers used when chaining formatters
}

// mtemplate: caching of parsed template files is all custom.
var parsedCache map[string]*Template

func init() {
	parsedCache = make(map[string]*Template)
}

func (parent *state) clone(data reflect.Value) *state {
	return &state{parent: parent, data: data, wr: parent.wr}
}

// New creates a new template with the specified formatter map (which
// may be nil) to define auxiliary functions for formatting variables.
func New(fmap FormatterMap) *Template {
	t := new(Template)
	t.fmap = fmap
	t.ldelim = lbrace
	t.rdelim = rbrace
	t.elems = NewElemlist()
	// mtemplate:
	t.childData = make(map[string]*bytes.Buffer)
	t.blockData = make(map[string]*bytes.Buffer)
	return t
}

// Report error and stop executing.  The line number must be provided explicitly.
func (t *Template) execError(st *state, line int, err string, args ...interface{}) {
	panic(&Error{line, fmt.Sprintf(err, args...)})
}

// Report error, panic to terminate parsing.
// The line number comes from the template state.
func (t *Template) parseError(err string, args ...interface{}) {
	panic(&Error{t.linenum, fmt.Sprintf(err, args...)})
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// -- Lexical analysis

// Is c a white space character?
func white(c uint8) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }

// Safely, does s[n:n+len(t)] == t?
func equal(s []byte, n int, t []byte) bool {
	b := s[n:]
	if len(t) > len(b) { // not enough space left for a match.
		return false
	}
	for i, c := range t {
		if c != b[i] {
			return false
		}
	}
	return true
}

// nextItem returns the next item from the input buffer.  If the returned
// item is empty, we are at EOF.  The item will be either a
// delimited string or a non-empty string between delimited
// strings. Tokens stop at (but include, if plain text) a newline.
// Action tokens on a line by themselves drop any space on
// either side, up to and including the newline.
func (t *Template) nextItem() []byte {
	startOfLine := t.p == 0 || t.buf[t.p-1] == '\n'
	start := t.p
	var i int
	newline := func() {
		t.linenum++
		i++
	}
	// Leading white space up to but not including newline
	for i = start; i < len(t.buf); i++ {
		if t.buf[i] == '\n' || !white(t.buf[i]) {
			break
		}
	}
	leadingSpace := i > start
	// What's left is nothing, newline, delimited string, or plain text
	switch {
	case i == len(t.buf):
		// EOF; nothing to do
	case t.buf[i] == '\n':
		newline()
	case equal(t.buf, i, t.ldelim):
		if t.buf[i+1] == '.' || t.buf[i+1] == '#' || t.buf[i+1] == '!' {
			left := i         // Start of left delimiter.
			right := -1       // Will be (immediately after) right delimiter.
			haveText := false // Delimiters contain text.
			i += len(t.ldelim)
			// Find the end of the action.
			for ; i < len(t.buf); i++ {
				if t.buf[i] == '\n' {
					break
				}
				if equal(t.buf, i, t.rdelim) {
					i += len(t.rdelim)
					right = i
					break
				}
				haveText = true
			}
			if right < 0 {
				t.parseError("unmatched opening delimiter")
				return nil
			}
			// Is this a special action (starts with '.' or '#') and the only thing on the line?
			if startOfLine && haveText {
				firstChar := t.buf[left+len(t.ldelim)]
				if firstChar == '.' || firstChar == '#' {
					// It's special and the first thing on the line. Is it the last?
					for j := right; j < len(t.buf) && white(t.buf[j]); j++ {
						if t.buf[j] == '\n' {
							// Yes it is. Drop the surrounding space and return the {.foo}
							t.linenum++
							t.p = j + 1
							return t.buf[left:right]
						}
					}
				}
			}
			// No it's not. If there's leading space, return that.
			if leadingSpace {
				// not trimming space: return leading white space if there is some.
				t.p = left
				return t.buf[start:left]
			}
			// Return the word, leave the trailing space.
			start = left

			break
		}

		fallthrough
	default:
		for ; i < len(t.buf); i++ {
			if t.buf[i] == '\n' {
				newline()
				break
			}
			if equal(t.buf, i, t.ldelim) && (t.buf[i+1] == '.' || t.buf[i+1] == '#' || t.buf[i+1] == '!') {
				break
			}
		}
	}
	item := t.buf[start:i]
	t.p = i
	return item
}

// Turn a byte array into a white-space-split array of strings.
func words(buf []byte) []string {
	s := make([]string, 0, 5)
	p := 0 // position in buf
	// one word per loop
	for i := 0; ; i++ {
		// skip white space
		for ; p < len(buf) && white(buf[p]); p++ {
		}
		// grab word
		start := p
		for ; p < len(buf) && !white(buf[p]); p++ {
		}
		if start == p { // no text left
			break
		}
		s = append(s, string(buf[start:p]))
	}
	return s
}

// Analyze an item and return its token type and, if it's an action item, an array of
// its constituent words.
func (t *Template) analyze(item []byte) (tok int, w []string) {
	// item is known to be non-empty
	if !equal(item, 0, t.ldelim) { // doesn't start with left delimiter
		tok = tokText
		return
	}
	if !equal(item, len(item)-len(t.rdelim), t.rdelim) { // doesn't end with right delimiter
		t.parseError("internal error: unmatched opening delimiter") // lexing should prevent this
		return
	}
	if len(item) <= len(t.ldelim)+len(t.rdelim) { // no contents
		t.parseError("empty directive")
		return
	}
	// Comment
	if item[len(t.ldelim)] == '#' {
		tok = tokComment
		return
	}
	// Split into words
	w = words(item[len(t.ldelim) : len(item)-len(t.rdelim)]) // drop final delimiter
	if len(w) == 0 {
		t.parseError("empty directive")
		return
	}
	if len(w) > 0 && w[0][0] == '!' {
		tok = tokVariable
		return
	}
	switch w[0] {
	case ".meta-left", ".meta-right", ".space", ".tab":
		tok = tokLiteral
		return
	case ".or":
		tok = tokOr
		return
	case ".end":
		tok = tokEnd
		return
	case ".section":
		if len(w) != 2 {
			t.parseError("incorrect fields for .section: %s", item)
			return
		}
		tok = tokSection
		return
	case ".repeated":
		if len(w) != 3 || w[1] != "section" {
			t.parseError("incorrect fields for .repeated: %s", item)
			return
		}
		tok = tokRepeated
		return
	case ".alternates":
		if len(w) != 2 || w[1] != "with" {
			t.parseError("incorrect fields for .alternates: %s", item)
			return
		}
		tok = tokAlternates
		return
		// mtemplate: parse .parent and .child nodes
	case ".parent":
		if len(w) < 2 {
			// It is possible that the parent template will be provided
			// when the caller chooses the ParseWithParent
			// function to create the template. In that case, it is
			// not an error to have a .parent command without
			// a parameter.
			if t.parent == nil || len(t.parent.body) == 0 {
				t.parseError("incorrect fields for .parent: %s", item)
				return
			}
		}
		tok = tokParent
		return
	case ".child":
		tok = tokChild
		return
	case ".block":
		if len(w) < 2 {
			t.parseError(".block must include a name")
			return
		}
		tok = tokBlock
		return
	case ".include":
		if len(w) < 2 {
			t.parseError(".include must specify the name of a file to include")
			return
		}
		tok = tokInclude
		return
	}

	t.parseError("bad directive: %s", item)
	return
}

// formatter returns the Formatter with the given name in the Template, or nil if none exists.
func (t *Template) formatter(name string) func(io.Writer, string, ...interface{}) {
	if t.fmap != nil {
		if fn := t.fmap[name]; fn != nil {
			return fn
		}
	}
	return builtins[name]
}

// -- Parsing

// Allocate a new variable-evaluation element.
func (t *Template) newVariable(words []string) *variableElement {
	// After the final space-separated argument, formatters may be specified separated
	// by pipe symbols, for example: {a b c|d|e}

	// Until we learn otherwise, formatters contains a single name: "", the default formatter.
	formatters := []string{""}
	lastWord := words[len(words)-1]
	bar := strings.IndexRune(lastWord, '|')
	if bar >= 0 {
		words[len(words)-1] = lastWord[0:bar]
		formatters = strings.Split(lastWord[bar+1:], "|")
	}

	// We could remember the function address here and avoid the lookup later,
	// but it's more dynamic to let the user change the map contents underfoot.
	// We do require the name to be present, though.

	// Is it in user-supplied map?
	for _, f := range formatters {
		if t.formatter(f) == nil {
			t.parseError("unknown formatter: %q", f)
		}
	}

	return &variableElement{t.linenum, words, formatters}
}

// Grab the next item.  If it's simple, just append it to the template.
// Otherwise return its details.
func (t *Template) parseSimple(item []byte, elems *elemlist) (done bool, tok int, w []string) {
	tok, w = t.analyze(item)
	done = true // assume for simplicity
	switch tok {
	case tokComment:
		return
	case tokText:
		elems.Push(&textElement{item})
		return
	case tokLiteral:
		switch w[0] {
		case ".meta-left":
			elems.Push(&literalElement{t.ldelim})
		case ".meta-right":
			elems.Push(&literalElement{t.rdelim})
		case ".space":
			elems.Push(&literalElement{space})
		case ".tab":
			elems.Push(&literalElement{tab})
		default:
			t.parseError("internal error: unknown literal: %s", w[0])
		}
		return
	case tokVariable:
		// Need to strip the leading ! off of our variable declaration.
		w[0] = w[0][1:]
		// Keep only non-whitespace elements
		for i := len(w) - 1; i >= 0; i-- {
			if len(strings.TrimSpace(w[i])) == 0 {
				w = append(w[:i], w[i+1:]...)
			}
		}
		elems.Push(t.newVariable(w))
		return
	}
	return false, tok, w
}

// parseRepeated and parseSection are mutually recursive

func (t *Template) parseRepeated(words []string, elems *elemlist) *repeatedElement {
	r := new(repeatedElement)
	elems.Push(r)
	r.linenum = t.linenum
	r.field = words[2]
	// Scan section, collecting true and false (.or) blocks.
	r.start = t.elems.Len()
	r.or = -1
	r.altstart = -1
	r.altend = -1
Loop:
	for {
		item := t.nextItem()
		if len(item) == 0 {
			t.parseError("missing .end for .repeated section")
			break
		}
		done, tok, w := t.parseSimple(item, elems)
		if done {
			continue
		}
		switch tok {
		case tokEnd:
			break Loop
		case tokOr:
			if r.or >= 0 {
				t.parseError("extra .or in .repeated section")
				break Loop
			}
			r.altend = elems.Len()
			r.or = elems.Len()
		case tokSection:
			t.parseSection(w, elems)
		case tokRepeated:
			t.parseRepeated(w, elems)
		case tokAlternates:
			if r.altstart >= 0 {
				t.parseError("extra .alternates in .repeated section")
				break Loop
			}
			if r.or >= 0 {
				t.parseError(".alternates inside .or block in .repeated section")
				break Loop
			}
			r.altstart = elems.Len()
		default:
			t.parseError("internal error: unknown repeated section item: %s", item)
			break Loop
		}
	}
	if r.altend < 0 {
		r.altend = elems.Len()
	}
	r.end = elems.Len()
	return r
}

func (t *Template) parseParentPage(words []string) {
	if t.parent == nil {
		t.parent = new(parentElement)
	}

	if len(words) > 1 {
		t.parent.filename = words[1]
	}
}

func (t *Template) parseChild(words []string, elems *elemlist) {
	newChild := new(childElement)
	if len(words) == 2 {
		newChild.name = words[1]
	}
	elems.Push(newChild)
}

func (t *Template) parseBlock(words []string, elems *elemlist) {
	if len(words) == 2 {
		newBlock := new(blockElement)
		newBlock.elems = NewElemlist()
		newBlock.name = words[1]
		t.parse(newBlock.elems, true)
		elems.Push(newBlock)
	}
}

func (t *Template) parseInclude(words []string, elems *elemlist) {
	if len(words) == 2 {
		newInclude := new(includeElement)
		newInclude.fileName = words[1]
		// By calling MustParseFile, we ensure that caching will be
		// used to provide this file. That could greatly improve performance.
		tempTemplate := MustParseFile(newInclude.fileName, nil)
		newInclude.elems = tempTemplate.elems
		elems.Push(newInclude)
	}
}

func (t *Template) parseSection(words []string, elems *elemlist) *sectionElement {
	s := new(sectionElement)
	elems.Push(s)
	s.linenum = t.linenum
	s.field = words[1]
	// Scan section, collecting true and false (.or) blocks.
	s.start = t.elems.Len()
	s.or = -1
Loop:
	for {
		item := t.nextItem()
		if len(item) == 0 {
			t.parseError("missing .end for .section")
			break
		}
		done, tok, w := t.parseSimple(item, elems)
		if done {
			continue
		}
		switch tok {
		case tokEnd:
			break Loop
		case tokOr:
			if s.or >= 0 {
				t.parseError("extra .or in .section")
				break Loop
			}
			s.or = elems.Len()
		case tokSection:
			t.parseSection(w, elems)
		case tokRepeated:
			t.parseRepeated(w, elems)
		case tokAlternates:
			t.parseError(".alternates not in .repeated")
		case tokInclude:
			t.parseInclude(w, elems)
		default:
			t.parseError("internal error: unknown section item: %s", item)
		}
	}
	s.end = elems.Len()
	return s
}

func (t *Template) parse(elems *elemlist, exitOnEnd bool) {
	for {
		item := t.nextItem()
		if len(item) == 0 {
			break
		}
		done, tok, w := t.parseSimple(item, elems)
		if done {
			continue
		}
		switch tok {
		case tokOr, tokAlternates:
			t.parseError("unexpected %s", w[0])
		case tokSection:
			t.parseSection(w, elems)
		case tokRepeated:
			t.parseRepeated(w, elems)
			// mtemplate:
		case tokParent:
			t.parseParentPage(w)
		case tokChild:
			t.parseChild(w, elems)
		case tokBlock:
			t.parseBlock(w, elems)
		case tokEnd:
			if exitOnEnd {
				return
			} else {
				t.parseError("unexpected %s", w[0])
			}
		case tokInclude:
			t.parseInclude(w, elems)
		default:
			t.parseError("internal error: bad directive in parse: %s", item)
		}
	}
}

// -- Execution

// Evaluate interfaces and pointers looking for a value that can look up the name, via a
// struct field, method, or map key, and return the result of the lookup.
func (t *Template) lookup(st *state, v reflect.Value, name string) reflect.Value {
	for v.IsValid() {
		typ := v.Type()
		if n := v.Type().NumMethod(); n > 0 {
			for i := 0; i < n; i++ {
				m := typ.Method(i)
				mtyp := m.Type
				if m.Name == name && mtyp.NumIn() == 1 && mtyp.NumOut() == 1 {
					if !isExported(name) {
						t.execError(st, t.linenum, "name not exported: %s in type %s", name, st.data.Type())
					}
					return v.Method(i).Call(nil)[0]
				}
			}
		}
		switch av := v; av.Kind() {
		case reflect.Ptr:
			v = av.Elem()
		case reflect.Interface:
			v = av.Elem()
		case reflect.Struct:
			if !isExported(name) {
				t.execError(st, t.linenum, "name not exported: %s in type %s", name, st.data.Type())
			}
			return av.FieldByName(name)
		case reflect.Map:
			if v := av.MapIndex(reflect.ValueOf(name)); v.IsValid() {
				return v
			}
			return reflect.Zero(typ.Elem())
		default:
			return reflect.Value{}
		}
	}
	return v
}

// indirectPtr returns the item numLevels levels of indirection below the value.
// It is forgiving: if the value is not a pointer, it returns it rather than giving
// an error.  If the pointer is nil, it is returned as is.
func indirectPtr(v reflect.Value, numLevels int) reflect.Value {
	for i := numLevels; v.IsValid() && i > 0; i++ {
		if p := v; p.Kind() == reflect.Ptr {
			if p.IsNil() {
				return v
			}
			v = p.Elem()
		} else {
			break
		}
	}
	return v
}

// Walk v through pointers and interfaces, extracting the elements within.
func indirect(v reflect.Value) reflect.Value {
loop:
	for v.IsValid() {
		switch av := v; av.Kind() {
		case reflect.Ptr:
			v = av.Elem()
		case reflect.Interface:
			v = av.Elem()
		default:
			break loop
		}
	}
	return v
}

// If the data for this template is a struct, find the named variable.
// Names of the form a.b.c are walked down the data tree.
// The special name "@" (the "cursor") denotes the current data.
// The value coming in (st.data) might need indirecting to reach
// a struct while the return value is not indirected - that is,
// it represents the actual named field. Leading stars indicate
// levels of indirection to be applied to the value.
func (t *Template) findVar(st *state, s string) reflect.Value {
	data := st.data
	flattenedName := strings.TrimLeft(s, "*")
	numStars := len(s) - len(flattenedName)
	s = flattenedName
	if s == "@" {
		return indirectPtr(data, numStars)
	}
	for _, elem := range strings.Split(s, ".") {
		// Look up field; data must be a struct or map.
		data = t.lookup(st, data, elem)
		if !data.IsValid() {
			return reflect.Value{}
		}
	}
	return indirectPtr(data, numStars)
}

// Is there no data to look at?
func empty(v reflect.Value) bool {
	v = indirect(v)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool() == false
	case reflect.String:
		return v.String() == ""
	case reflect.Struct:
		return false
	case reflect.Map:
		return false
	case reflect.Array:
		return v.Len() == 0
	case reflect.Slice:
		return v.Len() == 0
	}
	return false
}

// Look up a variable or method, up through the parent if necessary.
func (t *Template) varValue(name string, st *state) reflect.Value {
	field := t.findVar(st, name)
	if !field.IsValid() {
		if st.parent == nil {

			return reflect.Zero(reflect.TypeOf(""))
		}
		return t.varValue(name, st.parent)
	}
	return field
}

func (t *Template) format(wr io.Writer, fmt string, val []interface{}, v *variableElement, st *state) {
	fn := t.formatter(fmt)
	if fn == nil {
		t.execError(st, v.linenum, "missing formatter %s for variable %s", fmt, v.word[0])
	}
	fn(wr, fmt, val...)
}

// Evaluate a variable, looking up through the parent if necessary.
// If it has a formatter attached ({var|formatter}) run that too.
func (t *Template) writeVariable(v *variableElement, st *state) {
	// Turn the words of the invocation into values.
	
	val := make([]interface{}, len(v.word))
	for i, word := range v.word {
		val[i] = t.varValue(word, st).Interface()
	}

	for i, fmt := range v.fmts[:len(v.fmts)-1] {
		b := &st.buf[i&1]
		b.Reset()
		t.format(b, fmt, val, v, st)
		val = val[0:1]
		val[0] = b.Bytes()
	}
	t.format(st.wr, v.fmts[len(v.fmts)-1], val, v, st)
}

// Execute element i.  Return next index to execute.
func (t *Template) executeElement(elems *elemlist, i int, st *state) int {
	switch elem := elems.At(i).(type) {
	case *textElement:
		st.wr.Write(elem.text)
		return i + 1
	case *literalElement:
		st.wr.Write(elem.text)
		return i + 1
	case *variableElement:
		t.writeVariable(elem, st)
		return i + 1
	case *sectionElement:
		t.executeSection(elems, elem, st)
		return elem.end
	case *repeatedElement:
		t.executeRepeated(elems, elem, st)
		return elem.end
	// mtemplate:
	case *childElement:
		if blockBuffer, ok := t.childData[elem.name]; ok {
			blockBuffer.WriteTo(st.wr)
		}
		return i + 1
	case *blockElement:
		blockBuffer := new(bytes.Buffer)
		t.execute(elem.elems, &state{parent: st.parent, data: st.data, wr: blockBuffer})
		t.blockData[elem.name] = blockBuffer

		return i + 1
	case *includeElement:
		t.execute(elem.elems, st)
		return i + 1
	}
	e := t.elems.At(i)
	t.execError(st, 0, "internal error: bad directive in execute: %v %T\n", reflect.ValueOf(e).Interface(), e)
	return 0
}

// Execute the template.
func (t *Template) execute(elems *elemlist, st *state) {
	for i := 0; i < elems.Len(); {
		i = t.executeElement(elems, i, st)
	}
}

// Execute a .section
func (t *Template) executeSection(elems *elemlist, s *sectionElement, st *state) {
	// Find driver data for this section.  It must be in the current struct.
	field := t.varValue(s.field, st)
	if !field.IsValid() {
		t.execError(st, s.linenum, ".section: cannot find field %s in %s", s.field, st.data.Type())
	}
	st = st.clone(field)
	start, end := s.start, s.or
	if !empty(field) {
		// Execute the normal block.
		if end < 0 {
			end = s.end
		}
	} else {
		// Execute the .or block.  If it's missing, do nothing.
		start, end = s.or, s.end
		if start < 0 {
			return
		}
	}
	for i := start; i < end; {
		i = t.executeElement(elems, i, st)
	}
}

// Return the result of calling the Iter method on v, or nil.
func iter(v reflect.Value) reflect.Value {
	for j := 0; j < v.Type().NumMethod(); j++ {
		mth := v.Type().Method(j)
		fv := v.Method(j)
		ft := fv.Type()
		// TODO(rsc): NumIn() should return 0 here, because ft is from a curried FuncValue.
		if mth.Name != "Iter" || ft.NumIn() != 1 || ft.NumOut() != 1 {
			continue
		}
		ct := ft.Out(0)
		if ct.Kind() != reflect.Chan ||
			ct.ChanDir()&reflect.RecvDir == 0 {
			continue
		}
		return fv.Call(nil)[0]
	}
	return reflect.Value{}
}

// Execute a .repeated section
func (t *Template) executeRepeated(elems *elemlist, r *repeatedElement, st *state) {
	// Find driver data for this section.  It must be in the current struct.
	field := t.varValue(r.field, st)
	if !field.IsValid() {
		t.execError(st, r.linenum, ".repeated: cannot find field %s in %s", r.field, st.data.Type())
	}
	field = indirect(field)

	start, end := r.start, r.or
	if end < 0 {
		end = r.end
	}
	if r.altstart >= 0 {
		end = r.altstart
	}
	first := true

	// Code common to all the loops.
	loopBody := func(newst *state) {
		// .alternates between elements
		if !first && r.altstart >= 0 {
			for i := r.altstart; i < r.altend; {
				i = t.executeElement(elems, i, newst)
			}
		}
		first = false
		for i := start; i < end; {
			i = t.executeElement(elems, i, newst)
		}
	}

	if array := field; array.Kind() == reflect.Array || array.Kind() == reflect.Slice {
		for j := 0; j < array.Len(); j++ {
			loopBody(st.clone(array.Index(j)))
		}
	} else if m := field; m.Kind() == reflect.Map {
		for _, key := range m.MapKeys() {
			loopBody(st.clone(m.MapIndex(key)))
		}
	} else if ch := iter(field); ch.IsValid() {
		for {
			e, ok := ch.Recv()
			if !ok {
				break
			}
			loopBody(st.clone(e))
		}
	} else {
		t.execError(st, r.linenum, ".repeated: cannot repeat %s (type %s)",
			r.field, field.Type())
	}

	if first {
		// Empty. Execute the .or block, once.  If it's missing, do nothing.
		start, end := r.or, r.end
		if start >= 0 {
			newst := st.clone(field)
			for i := start; i < end; {
				i = t.executeElement(elems, i, newst)
			}
		}
		return
	}
}

// A valid delimiter must contain no white space and be non-empty.
func validDelim(d []byte) bool {
	if len(d) == 0 {
		return false
	}
	for _, c := range d {
		if white(c) {
			return false
		}
	}
	return true
}

// checkError is a deferred function to turn a panic with type *Error into a plain error return.
// Other panics are unexpected and so are re-enabled.
func checkError(error *error) {
	if v := recover(); v != nil {
		if e, ok := v.(*Error); ok {
			*error = e
		} else {
			// runtime errors should crash
			panic(v)
		}
	}
}

// -- Public interface

// Parse initializes a Template by parsing its definition.  The string
// s contains the template text.  If any errors occur, Parse returns
// the error.
func (t *Template) Parse(s string) (err error) {
	if t.elems == nil {
		return &Error{1, "template not allocated with New"}
	}
	if !validDelim(t.ldelim) || !validDelim(t.rdelim) {
		return &Error{1, fmt.Sprintf("bad delimiter strings %q %q", t.ldelim, t.rdelim)}
	}
	defer checkError(&err)
	t.buf = []byte(s)
	t.p = 0
	t.linenum = 1
	t.parse(t.elems, false)
	return nil
}

// ParseFile is like Parse but reads the template definition from the
// named file.
func (t *Template) ParseFile(filename string) (err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return t.Parse(string(b))
}

// Execute applies a parsed template to the specified data object,
// generating output to wr.
func (t *Template) Execute(wr io.Writer, data interface{}) (err error) {
	// Extract the driver data.
	val := reflect.ValueOf(data)
	defer checkError(&err)
	t.p = 0
	// mtemplate: parent/child-specific functionality
	if t.parent != nil {
		childDocWriter := new(bytes.Buffer)
		t.execute(t.elems, &state{parent: nil, data: val, wr: childDocWriter})

		// Now, process the Parent
		var parentTemplate *Template

		if len(t.parent.filename) > 0 {
			parentTemplate = MustParseFile(t.parent.filename, nil)
		} else {
			parentTemplate = MustParse(t.parent.body, nil)
		}
		parentTemplateCopy := parentTemplate.clone()
		parentTemplateCopy.childData = t.blockData
		parentTemplateCopy.childData[""] = childDocWriter
		parentTemplateCopy.Execute(wr, data)
	} else {
		// mtemplate: this code handles templates that do not use
		// the parent/child functionality
		t.execute(t.elems, &state{parent: nil, data: val, wr: wr})
	}

	return nil
}

// SetDelims sets the left and right delimiters for operations in the
// template.  They are validated during parsing.  They could be
// validated here but it's better to keep the routine simple.  The
// delimiters are very rarely invalid and Parse has the necessary
// error-handling interface already.
func (t *Template) SetDelims(left, right string) {
	t.ldelim = []byte(left)
	t.rdelim = []byte(right)
}

// Parse creates a Template with default parameters (such as {} for
// metacharacters).  The string s contains the template text while
// the formatter map fmap, which may be nil, defines auxiliary functions
// for formatting variables.  The template is returned. If any errors
// occur, err will be non-nil.
func Parse(s string, fmap FormatterMap) (t *Template, err error) {
	t = New(fmap)
	err = t.Parse(s)
	if err != nil {
		t = nil
	}
	return
}

// ParseWithParentPage is used to parse a template that has a
// .parent directive when both the template and the parent
// template must be supplied as strings rather than being read off of 
// the filesystem. This might not have a practical use case aside
// from unit-testing, but it does seem to be required to avoid
// unit tests that create lots of temporary files.
func ParseWithParentPage(s string, parentBody string, fmap FormatterMap) (t *Template, err error) {
	// mtemplate:
	t = New(fmap)
	t.parent = new(parentElement)
	t.parent.body = parentBody
	err = t.Parse(s)
	if err != nil {
		t = nil
	}
	return
}

// ParseFile is a wrapper function that creates a Template with default
// parameters (such as {} for metacharacters).  The filename identifies
// a file containing the template text, while the formatter map fmap, which
// may be nil, defines auxiliary functions for formatting variables.
// The template is returned. If any errors occur, err will be non-nil.
func ParseFile(filename string, fmap FormatterMap) (t *Template, err error) {
	template, ok := parsedCache[filename]
	if !ok {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		t, err = Parse(string(b), fmap)
		parsedCache[filename] = t
	} else {
		t = template
	}

	return
}

// MustParse is like Parse but panics if the template cannot be parsed.
func MustParse(s string, fmap FormatterMap) *Template {
	t, err := Parse(s, fmap)
	if err != nil {
		panic("template.MustParse error: " + err.Error())
	}
	return t
}

// MustParseFile is like ParseFile but panics if the file cannot be read
// or the template cannot be parsed.
func MustParseFile(filename string, fmap FormatterMap) *Template {
	template, ok := parsedCache[filename]
	if !ok {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			panic("template.MustParseFile error: " + err.Error())
		}
		template = MustParse(string(b), fmap)
		parsedCache[filename] = template
	}

	return template
}

func ClearFromCache(filename string) {
	delete(parsedCache, filename)
}

func RenderFile(filename string, wr io.Writer, data interface{}) (err error) {
	template, err := ParseFile(filename, nil)
	if err != nil {
		return err
	}

	template.Execute(wr, data)
	return
}

func MustRenderFile(filename string, wr io.Writer, data interface{}) {
	template := MustParseFile(filename, nil)

	err := template.Execute(wr, data)
	if err != nil {
		panic("template.MustRenderFile error executing template: " + err.Error())
	}

	return
}

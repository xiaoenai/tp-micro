// Package html is HTML render for http client.
package html

import (
	"sync"
	"text/template"
	"time"

	"github.com/henrylee2cn/erpc/v6"
	"github.com/henrylee2cn/erpc/v6/utils"
	micro "github.com/xiaoenai/tp-micro/v6"
)

// CtxMeta the metadata method sets of context
type CtxMeta interface {
	// SetMeta sets the header metadata 'key=value' for reply packet.
	SetMeta(key, value string)
}

var (
	t                = template.New("")
	tLocker          sync.RWMutex
	parseFuncs       = make(map[string]func(*template.Template), 128)
	parseFuncsLocker sync.RWMutex
)

// Delims sets the action delimiters to the specified strings, to be used in
// subsequent calls to Parse, ParseFiles, or ParseGlob. Nested template
// definitions will inherit the settings. An empty delimiter stands for the
// corresponding default: {{ or }}.
// The return value is the template, so calls can be chained.
//
// Note: Must be called before Parse, ParseFiles, and ParseGlob.
//
func Delims(left, right string) {
	tLocker.Lock()
	defer tLocker.Unlock()
	t.Delims(left, right)
}

// Funcs adds the elements of the argument map to the template's function map.
// It must be called before the template is parsed.
// It panics if a value in the map is not a function with appropriate return
// type or if the name cannot be used syntactically as a function in a template.
// It is legal to overwrite elements of the map. The return value is the template,
// so calls can be chained.
func Funcs(funcMap template.FuncMap) {
	tLocker.Lock()
	defer tLocker.Unlock()
	t.Funcs(funcMap)
}

// Option sets options for the template. Options are described by
// strings, either a simple string or "key=value". There can be at
// most one equals sign in an option string. If the option string
// is unrecognized or otherwise invalid, Option panics.
//
// Known options:
//
// missingkey: Control the behavior during execution if a map is
// indexed with a key that is not present in the map.
//	"missingkey=default" or "missingkey=invalid"
//		The default behavior: Do nothing and continue execution.
//		If printed, the result of the index operation is the string
//		"<no value>".
//	"missingkey=zero"
//		The operation returns the zero value for the map type's element.
//	"missingkey=error"
//		Execution stops immediately with an error.
//
func Option(opt ...string) {
	tLocker.Lock()
	defer tLocker.Unlock()
	t.Option(opt...)
}

// Parse parses text as a template body for t.
// Named template definitions ({{define ...}} or {{block ...}} statements) in text
// define additional templates associated with t and are removed from the
// definition of t itself.
//
// Templates can be redefined in successive calls to Parse.
// A template definition with a body containing only white space and comments
// is considered empty and will not replace an existing template's body.
// This allows using Parse to add new named template definitions without
// overwriting the main template body.
func Parse(name, text string) error {
	tLocker.Lock()
	defer tLocker.Unlock()
	_, err := t.New(name).Parse(text)
	return err
}

// ParseFiles parses the named files and associates the resulting templates with
// t. If an error occurs, parsing stops and the returned template is nil;
// otherwise it is t. There must be at least one file.
// Since the templates created by ParseFiles are named by the base
// names of the argument files, t should usually have the name of one
// of the (base) names of the files. If it does not, depending on t's
// contents before calling ParseFiles, t.Execute may fail. In that
// case use t.ExecuteTemplate to execute a valid template.
//
// When parsing multiple files with the same name in different directories,
// the last one mentioned will be the one that results.
func ParseFiles(filenames ...string) (err error) {
	tLocker.Lock()
	defer func() {
		tLocker.Unlock()
		if err != nil {
			return
		}
		parseFuncsLocker.Lock()
		defer parseFuncsLocker.Unlock()
		for _, f := range filenames {
			parseFuncs[f] = func(nt *template.Template) {
				if _, err := nt.ParseFiles(f); err != nil {
					erpc.Errorf("ParseFiles: filename: %s, error: %s", f, err.Error())
				}
			}
		}
	}()
	_, err = t.ParseFiles(filenames...)
	return err
}

// ParseGlob creates a new Template and parses the template definitions from the
// files identified by the pattern, which must match at least one file. The
// returned template will have the (base) name and (parsed) contents of the
// first file matched by the pattern. ParseGlob is equivalent to calling
// ParseFiles with the list of files matched by the pattern.
//
// When parsing multiple files with the same name in different directories,
// the last one mentioned will be the one that results.
//
// The pattern syntax is:
//
//	pattern:
//		{ term }
//	term:
//		'*'         matches any sequence of non-Separator characters
//		'?'         matches any single non-Separator character
//		'[' [ '^' ] { character-range } ']'
//		            character class (must be non-empty)
//		c           matches character c (c != '*', '?', '\\', '[')
//		'\\' c      matches character c
//
//	character-range:
//		c           matches character c (c != '\\', '-', ']')
//		'\\' c      matches character c
//		lo '-' hi   matches character c for lo <= c <= hi
//
// Match requires pattern to match all of name, not just a substring.
// The only possible returned error is ErrBadPattern, when pattern
// is malformed.
//
// On Windows, escaping is disabled. Instead, '\\' is treated as
// path separator.
//
func ParseGlob(pattern string) (err error) {
	tLocker.Lock()
	defer func() {
		tLocker.Unlock()
		if err != nil {
			return
		}
		parseFuncsLocker.Lock()
		defer parseFuncsLocker.Unlock()
		parseFuncs[pattern] = func(nt *template.Template) {
			if _, err := nt.ParseGlob(pattern); err != nil {
				erpc.Errorf("ParseGlob: pattern: %s, error: %s", pattern, err.Error())
			}
		}
	}()
	_, err = t.ParseGlob(pattern)
	return err
}

// GoTimingRefresh runs a goroutine that periodically refreshes template files.
func GoTimingRefresh(d time.Duration) {
	go func() {
		ticker := time.NewTicker(d)
		for range ticker.C {
			parseFuncsLocker.RLock()
			nt, _ := t.Clone()
			for _, fn := range parseFuncs {
				fn(nt)
			}
			parseFuncsLocker.RUnlock()
			tLocker.Lock()
			t = nt
			tLocker.Unlock()
		}
	}()
}

// Render renders the initialized html template by name.
func Render(ctxMeta CtxMeta, tmplName string, data interface{}) ([]byte, *erpc.Status) {
	ctxMeta.SetMeta("Content-Type", "text/html; charset=utf-8")
	buf := utils.AcquireByteBuffer()
	tLocker.RLock()
	err := t.ExecuteTemplate(buf, tmplName, data)
	tLocker.RUnlock()
	if err != nil {
		return nil, micro.RerrRenderFailed.Copy(err)
	}
	return buf.Bytes(), nil
}

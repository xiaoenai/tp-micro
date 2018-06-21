## html

HTML render for http client.

### Usage

`html "github.com/xiaoenai/tp-micro/helper/mod-html"`

#### Test

```go
package html_test

import (
	"testing"
	"time"

	html "github.com/xiaoenai/tp-micro/helper/mod-html"
)

type meta struct{}

func (m *meta) SetMeta(key, value string) {}

func TestParseText(t *testing.T) {
	const (
		doc1 = `<!DOCTYPE html>
	<html>
	<head>
	    <title>html test 1</title>
	</head>
	<body>
	<br/>
	<h2><center>{{.}}</center></h2>
	</body>
	</html>`
		doc2 = `<!DOCTYPE html>
	<html>
	<head>
	    <title>html test 2</title>
	</head>
	<body>
	<br/>
	<h2><center>{{.}}</center></h2>
	</body>
	</html>`
	)

	err := html.Parse("a", doc1)
	if err != nil {
		t.Fatal(err)
	}
	err = html.Parse("b", doc2)
	if err != nil {
		t.Fatal(err)
	}

	b, rerr := html.Render(new(meta), "a", "TestParse A!")
	t.Logf("body: %s, rerr: %v", b, rerr)
	b, rerr = html.Render(new(meta), "b", "TestParse B!")
	t.Logf("body: %s, rerr: %v", b, rerr)
}

func TestParseFiles(t *testing.T) {
	err := html.ParseFiles("../mod-html/a_test.tpl", "../mod-html/b_test.tpl")
	if err != nil {
		t.Fatal(err)
	}

	b, rerr := html.Render(new(meta), "a_test.tpl", "TestParseFiles A!")
	t.Logf("body: %s, rerr: %v", b, rerr)
	b, rerr = html.Render(new(meta), "b_test.tpl", "TestParseFiles B!")
	t.Logf("body: %s, rerr: %v", b, rerr)
}

func TestParseGlob(t *testing.T) {
	err := html.ParseGlob("../mod-html/*.tpl")
	if err != nil {
		t.Fatal(err)
	}

	b, rerr := html.Render(new(meta), "a_test.tpl", "TestParseGlob A!")
	t.Logf("body: %s, rerr: %v", b, rerr)
	b, rerr = html.Render(new(meta), "b_test.tpl", "TestParseGlob B!")
	t.Logf("body: %s, rerr: %v", b, rerr)
}

func TestDelims(t *testing.T) {
	const (
		doc = `<!DOCTYPE html>
	<html>
	<head>
	    <title>html test 1</title>
	</head>
	<body>
	<br/>
	<h2><center>{{{.}}}</center></h2>
	</body>
	</html>`
	)

	html.Delims("{{{", "}}}")
	err := html.Parse("doc", doc)
	if err != nil {
		t.Fatal(err)
	}
	b, rerr := html.Render(new(meta), "doc", "TestDelims!")
	t.Logf("body: %s, rerr: %v", b, rerr)
}

func TestGoTimingRefresh(t *testing.T) {
	err := html.ParseFiles("../mod-html/a_test.tpl")
	if err != nil {
		t.Fatal(err)
	}
	err = html.ParseGlob("../mod-html/*.tpl")
	if err != nil {
		t.Fatal(err)
	}

	b, rerr := html.Render(new(meta), "a_test.tpl", "TestGoTimingRefresh 1!")
	t.Logf("test1: body: %s, rerr: %v", b, rerr)

	html.GoTimingRefresh(time.Second * 5)
	time.Sleep(time.Second * 15)

	b, rerr = html.Render(new(meta), "a_test.tpl", "TestGoTimingRefresh 2!")
	t.Logf("test2: body: %s, rerr: %v", b, rerr)
}
```

test command:

```sh
go test -v -run=TestParseText
go test -v -run=TestParseFiles
go test -v -run=TestParseGlob
go test -v -run=TestDelims
go test -v -run=TestGoTimingRefresh
```

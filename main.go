package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/russross/blackfriday"
)

const tpl = `
<html>
<head>
<style>
{{.css}}
</style>
</head>

<body>
<article class="markdown-body">
{{.body}}
</article>
</body>

</html>
`

func main() {
	flag.Parse()

	var r io.Reader = os.Stdin

	if flag.NArg() >= 1 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = f
	}

	input, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	body := blackfriday.MarkdownCommon(input)

	m := map[string]interface{}{
		"css":  css,
		"body": string(body),
	}
	template.Must(template.New("markdown").Parse(tpl)).Execute(os.Stdout, m)
}

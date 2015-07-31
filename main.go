package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

var (
	asServer = flag.Bool("server", false, "run as server")
	addr     = flag.String("addr", ":8080", "server address")
	root     = flag.String("root", ".", "server root")
)

func Markdown(in io.Reader, out io.Writer) error {
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	body := blackfriday.MarkdownCommon(buf)
	m := map[string]interface{}{
		"css":  css,
		"body": string(body),
	}
	return template.Must(template.New("markdown").Parse(tpl)).Execute(out, m)
}

func runcli() {
	var r io.Reader = os.Stdin

	if flag.NArg() >= 1 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = f
	}

	err := Markdown(r, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func serveMarkdown(w http.ResponseWriter, r *http.Request) {
	var code int = 200
	var err error
	defer func() {
		log.Printf("%s %d %s", r.Method, code, r.URL.Path)
		if err != nil {
			w.WriteHeader(code)
			io.WriteString(w, err.Error())
		}
	}()
	file := filepath.Join(*root, r.URL.Path)
	if !(strings.HasSuffix(file, ".md") || strings.HasSuffix(file, ".markdown")) {
		http.FileServer(http.Dir(*root)).ServeHTTP(w, r)
		return
	}
	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			code = 404
		}
		return
	}
	defer f.Close()
	err = Markdown(f, w)
	if err != nil {
		code = 500
		return
	}
}

func runserver() {
	log.Printf("Listening on %s, root %s", *addr, *root)
	http.HandleFunc("/", serveMarkdown)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func main() {
	flag.Parse()
	if !*asServer {
		runcli()
	} else {
		runserver()
	}
}

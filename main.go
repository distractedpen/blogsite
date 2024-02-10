package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/a-h/templ"
	"github.com/distractedpen/blogsite/pages"
	"github.com/distractedpen/blogsite/utils"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
)

//go:generate templ generate
//go:generate npm run build

func unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

func mdToHTML(md []byte) []byte {
	// setup parser
	extensions := parser.CommonExtensions | parser.Attributes | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func main() {

	const contentRoot = "content"
	const contentURL = "/content/{category}/{article}"

	r := mux.NewRouter()

    fs := http.FileServer(http.Dir("./static/"))
    r.Handle("/static/", http.StripPrefix("/static/", fs))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Print("root route: " + r.URL.Path)
        articles := utils.GetArticles()
        pages.IndexPage(articles).Render(context.Background(), w)
	})

	r.HandleFunc(contentURL, func(w http.ResponseWriter, r *http.Request) {
        log.Print("content route: " + r.URL.Path)
		vars := mux.Vars(r)
		category := vars["category"]
		article := vars["article"]

		articlePath := path.Join(contentRoot, category, article)

		source, read_err := os.ReadFile(articlePath + ".md")
		if read_err != nil {
			log.Print(read_err)
			pages.ErrorPage().Render(context.Background(), w)
			return
		}
		html := mdToHTML(source)

		w.Header().Set("Content-Type", "text/html")
		pages.ArticlePage(article, unsafe(string(html))).Render(context.Background(), w)
	})

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
    log.Print("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}

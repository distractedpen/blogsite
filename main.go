package main

import (
	"log"
    "os"
    "path"
	"net/http"
    "html/template"
	"time"

    "github.com/distractedpen/blogsite/utils"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:generate npm run build

func mdToHTML(source []byte) string{
    // setup parser
    extensions := parser.CommonExtensions | parser.Attributes | parser.AutoHeadingIDs
    p := parser.NewWithExtensions(extensions)
    doc := p.Parse(source)

    // create HTML
    htmlFlags := html.CommonFlags | html.HrefTargetBlank
    opts := html.RendererOptions{Flags: htmlFlags}
    renderer := html.NewRenderer(opts)

    return string(markdown.Render(doc, renderer))
}

func check(err error) {
    if err != nil {
        log.Panic(err)
    }
}

func main() {

	const contentRoot = "content"

    indexTemplateList := []string{
        "./templates/index.html",
        "./templates/components.html",
    }

    articleTemplateList := []string{
        "./templates/article.html",
        "./templates/components.html",
    }


    indexTemplates, err := template.ParseFiles(indexTemplateList...)
    check(err)

    articleTemplates, err := template.New("article.html").Funcs(template.FuncMap{
    }).ParseFiles(articleTemplateList...)

    log.Println(articleTemplates.DefinedTemplates())

	r := http.NewServeMux()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Print("root route: " + r.URL.Path)
        articleList := utils.ArticleList{ Articles: utils.GetArticles() }
        err := indexTemplates.Execute(w, articleList) 
        check(err)
	})

    fs := http.FileServer(http.Dir("./static/"))
    r.Handle("/static/", http.StripPrefix("/static/", fs))


	r.HandleFunc("/content/{category}/{article}", func(w http.ResponseWriter, r *http.Request) {
        category := r.PathValue("category")
        articleTitle := r.PathValue("article")

		articlePath := path.Join(contentRoot, category, articleTitle)

		source, read_err := os.ReadFile(articlePath + ".md")
		if read_err != nil {
			log.Print(read_err)
            http.Error(w, "Page not Found.", http.StatusInternalServerError)
			return
		}

        htmlContent := mdToHTML(source)
        article := utils.GetArticle(articlePath)

        articleContent := map[string]interface{}{
            "Title": article.Title,
            "DatePublished": article.DatePublished,
            "Content": template.HTML(htmlContent),
        }

        err := articleTemplates.Execute(w, articleContent)
        if (err != nil) {
            log.Panic(err)
        }
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
    log.Print("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}

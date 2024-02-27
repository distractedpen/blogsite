package main

import (
    "os"
    "fmt"
    "bufio"
    "strings"
    "path"
    "path/filepath"
    "html/template"

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
        panic(err)
    }
}

func main() {
    buildArticles()
    generateIndex()
    copyStaticFiles()
}

func copyStaticFiles() {
    // # Static File Copy
    // Copy all files from /static to /public/static
    staticFiles, err := os.ReadDir("static")
    check(err)

    for _, file := range staticFiles {
        if (file.IsDir()) {
            err := os.MkdirAll(path.Join("public", "static", file.Name()), 0775)
            check(err)
            filepath.Walk(path.Join("static", file.Name()), 
                func(fpath string, info os.FileInfo, err error) error {
                    check(err)
                    if (!info.IsDir() && info.Name() != "tailwind.css") {
                        source, read_err := os.ReadFile(fpath)
                        check(read_err)
                        dest := path.Join("public", "static", file.Name(), info.Name())
                        write_err := os.WriteFile(dest, source, 0775)
                        check(write_err)
                    }
                    return nil
                },
            )
        } else {
            source, read_err := os.ReadFile(path.Join("static", file.Name()))
            check(read_err)
            dest := path.Join("public", "static", file.Name())
            write_err := os.WriteFile(dest, source, 0775)
            check(write_err)
        }
    }
}

func generateIndex() {
    
    indexTemplateList := []string{
        "./templates/index.html",
        "./templates/components.html",
    }

    indexTemplates, err := template.ParseFiles(indexTemplateList...)
    check(err)

    // # Index Generation
    // Get File Tree of metadata
    var articleList []utils.Article
    metadataDir, err := os.ReadDir("metadata")
    check(err)
    // For each .yml file,
    for _, file := range metadataDir {
        articleList = append(articleList, utils.GetArticle(path.Join("metadata", file.Name())))
    }
    fmt.Println(articleList)

    // write baked template html to index.html
    f, err := os.Create("public/index.html")
    defer f.Close()

    //  write baked template html to /article
    w := bufio.NewWriter(f)
    indexWriteErr := indexTemplates.Execute(w, articleList)
    check(indexWriteErr)
}

func buildArticles() {

    articleTemplateList := []string{
        "./templates/article.html",
        "./templates/components.html",
    }
    // # Article Generation and Metadata Extraction
    // Create/Clear the output directory 
    err := os.Mkdir("public", 0775)
    if (err != nil) {
        if (os.IsExist(err)) {
            os.RemoveAll("public")
            os.Mkdir("public", 0775)
        } else {
            panic(err)
        }
    }

    err2 := os.Mkdir("metadata", 0775)
    if (err2 != nil) {
        if (os.IsExist(err2)) {
            os.RemoveAll("metadata")
            os.Mkdir("metadata", 0775)
        } else {
            panic(err2)
        }
    }

    // Get File Tree of content (given as arg later)
    filepath.Walk("content", 
        func(fpath string, info os.FileInfo, err error) error {
            check(err)

            if (info.IsDir()) {
                if (strings.Contains(fpath, "assets")) {
                    assetsPath := path.Join("public", strings.ReplaceAll(strings.TrimPrefix(fpath, "content/"), "/", "-"))
                    err := os.Mkdir(assetsPath, 0775)
                    check(err)
                    filepath.Walk(fpath, func(fpath string, info os.FileInfo, err error) error {
                        check(err)
                        if (!info.IsDir()) {
                            source, read_err := os.ReadFile(fpath)
                            check(read_err)
                            write_err := os.WriteFile(path.Join(assetsPath, info.Name()), source, 0775)
                            check(write_err)
                        }
                        return nil
                    })
                }
            }

            if (!strings.HasSuffix(fpath, ".md")) {
                return nil
            }

            articleFileName := strings.TrimPrefix(fpath, "content/")
            articleFileName = strings.TrimSuffix(strings.ReplaceAll(articleFileName, "/", "-"), ".md")
            //  read file
            source, read_err := os.ReadFile(fpath)
            check(read_err)

            //  find front matter
            foundFrontMatter := string(source[0:3]) == "---" // frontmatter exists
            frontMatter := ""
            endFrontMatter := 0
            if (foundFrontMatter) {
                for index, line := range strings.Split(string(source), "\n")[1:] {
                    if (line == "---") {
                        endFrontMatter = index
                        break
                    }
                    frontMatter += line + "\n"
                }
            }

            // strip front matter from source
            source = []byte(strings.Join(strings.Split(string(source), "\n")[endFrontMatter+1:], "\n"))

            //  write front matter as yml in /metadata . filename.yml
            metadataPath := path.Join("metadata", articleFileName + ".yml")
            metadata_write_err := os.WriteFile(metadataPath, []byte(frontMatter), 0775)
            check(metadata_write_err)
            //  convert rest to html
            htmlContent := mdToHTML(source)
            //  embed html into article template
            article := utils.GetArticle(fpath)
            articleContent := map[string]interface{}{
                "Title": article.ArticleMetadata.Title,
                "DatePublished": article.ArticleMetadata.DatePublished,
                "Content": template.HTML(htmlContent),
            }
            articleTemplates, err := template.New("article.html").Funcs(template.FuncMap{}).ParseFiles(articleTemplateList...)

            f, err := os.Create("public/" + articleFileName + ".html")
            defer f.Close()

            //  write baked template html to /article
            w := bufio.NewWriter(f)
            articleWriteErr := articleTemplates.Execute(w, articleContent)
            check(articleWriteErr)
            
            return nil
        },
    )
}

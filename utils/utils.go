package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Article struct {
	Title         string
	DatePublished time.Time
	ContentPath   string
}

type ArticleList struct {
	Articles []Article
}

type ArticleContent struct {
	Article Article
	Content []byte
}

func GetArticle(path string) Article {
	return buildArticle(path)
}

func GetArticles() []Article {

	articles := make([]Article, 0)

	// Read the content folder and generate list of article objects

	err := filepath.Walk("content",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

            article := buildArticle(path)
			articles = append(articles, article)
			return nil
		})

	if err != nil {
		log.Fatal(err)
	}

	return articles
}

func buildArticle(path string) Article {
	// Name Pattern: [date]_[title].md
	// Example: 2021-01-01_First-Article.md
	splitPath := strings.Split(path, "/")
	fileName := splitPath[len(splitPath)-1]

	noExt, _ := strings.CutSuffix(fileName, ".md")
	pathNoExt, _ := strings.CutSuffix(path, ".md")
	parts := strings.Split(noExt, "_")

	dateStr := parts[0]
	dateVal, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Invalid date format: %s", dateStr)
		dateVal = time.Now()
	}

	title := parts[1]

	return Article{
		Title:         title,
		DatePublished: dateVal,
		ContentPath:   pathNoExt,
	}
}


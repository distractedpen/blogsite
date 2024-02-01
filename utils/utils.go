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

			rawFileName := info.Name()
			// Name Pattern: [date]_[title].md
			// Example: 2021-01-01_First-Article.md
			noExt, _ := strings.CutSuffix(rawFileName, ".md")
            pathNoExt, _ := strings.CutSuffix(path, ".md")
			parts := strings.Split(noExt, "_")

			if len(parts) != 2 {
				log.Printf("Invalid file name: %s", rawFileName)
				return nil
			}

			dateStr := parts[0]
			dateVal, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Printf("Invalid date format: %s", dateStr)
				dateVal = time.Now()
				return nil
			}
			title := parts[1]

			article := Article{
				Title:         title,
				DatePublished: dateVal,
				ContentPath:   pathNoExt,
			}
			articles = append(articles, article)
			return nil
		})

    if err != nil {
        log.Fatal(err)
    }

	return articles
}

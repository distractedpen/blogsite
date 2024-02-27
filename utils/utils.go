package utils

import (
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ArticleMetadata struct {
	Title         string    `yaml:"title"`
	DatePublished time.Time `yaml:"datePublished"`
	Abstract      string    `yaml:"abstract"`
	Tags          []string  `yaml:"tags"`
}

type Article struct {
	ArticleMetadata ArticleMetadata
	Path            string
}

type ArticleList struct {
	Articles []Article
}

type ArticleContent struct {
	Article Article
	Content []byte
}

func GetArticle(fpath string) Article {
	// read the file and return the article object

	var metadataPath string
	if !strings.HasPrefix(fpath, "metadata") {
        metadataPath = path.Join("metadata", strings.TrimSuffix(strings.TrimPrefix(fpath, "content/"), ".md")+".yml")
	} else {
		metadataPath = fpath
	}

	articleMetadata, read_err := os.ReadFile(metadataPath)
	if read_err != nil {
		panic(read_err)
	}
	var metadata ArticleMetadata
	yaml_err := yaml.Unmarshal(articleMetadata, &metadata)
	if yaml_err != nil {
		panic(yaml_err)
	}

	return Article{
		ArticleMetadata: metadata,
		Path: strings.TrimPrefix(strings.Replace(fpath, "yml", "html", 1), "metadata"),
	}
}

func GetArticles() []Article {
	var articles []Article
	// Read the content folder and generate list of article objects
	err := filepath.Walk("metadata",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			article := GetArticle(path)
			articles = append(articles, article)
			return nil
		})

	if err != nil {
		panic(err)
	}

	return articles
}

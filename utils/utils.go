package utils

import (
	"time"
)

type Article struct {
	Title         string
	DatePublished time.Time
	ContentPath   string
}

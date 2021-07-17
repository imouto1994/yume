package model

type ScanResult struct {
	TitleByTitleName map[string]*Title
	BooksByTitleName map[string][]*Book
}

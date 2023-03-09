package models

import "gorm.io/gorm"

type Book struct {
	gorm.Model

	Slug        string `gorm:"uniqueIndex;not null"`
	Title       string `gorm:"uniqueIndex;not nul"`
	Description *string

	UploadedBy   User
	UploadedByID uint

	Reviews []Review

	Image         *string
	File          *string
	DownloadCount int

	Tags []Tag `gorm:"many2many:book_tags;"`
}

type Review struct {
	gorm.Model

	Book   Book
	BookID uint

	User   User
	UserID uint

	Title string
	Body  string
	Score int
}

type Tag struct {
	gorm.Model
	Tag   string `gorm:"uniqueIndex"`
	Books []Book `gorm:"many2many:book_tags;"`
}

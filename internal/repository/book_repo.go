package repository

import (
	"errors"

	"github.com/artqqwr/bookslib/internal/models"
	"gorm.io/gorm"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (repo *BookRepository) GetBySlug(s string) (*models.Book, error) {
	var m models.Book
	err := repo.db.Where(&models.Book{Slug: s}).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag asc")
		}).
		Preload("Author").First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, err
}

func (repo *BookRepository) GetUserArticleBySlug(userID uint, slug string) (*models.Book, error) {
	var m models.Book
	err := repo.db.Where(&models.Book{Slug: slug, UploadedByID: userID}).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &m, err
}

func (repo *BookRepository) CreateBook(a *models.Book) error {
	tags := a.Tags
	tx := repo.db.Begin()
	if err := tx.Create(&a).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, t := range a.Tags {
		err := tx.Where(&models.Tag{Tag: t.Tag}).First(&t).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}
		if err := tx.Model(&a).Association("Tags").Append(&t); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Where(a.ID).Preload("Tags").Preload("Author").First(&a).Error; err != nil {
		tx.Rollback()
		return err
	}
	a.Tags = tags
	return tx.Commit().Error
}

func (repo *BookRepository) UpdateBook(a *models.Book, tagList []string) error {
	tx := repo.db.Begin()
	if err := tx.Model(a).Updates(a).Error; err != nil {
		tx.Rollback()
		return err
	}
	tags := make([]models.Tag, 0)
	for _, t := range tagList {
		tag := models.Tag{Tag: t}

		err := tx.Where(&tag).First(&tag).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}
		tags = append(tags, tag)
	}
	if err := tx.Model(a).Association("Tags").Replace(tags); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where(a.ID).Preload("Tags").Preload("Author").First(a).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repo *BookRepository) DeleteBook(a *models.Book) error {
	return repo.db.Delete(a).Error
}

func (repo *BookRepository) List(offset, limit int) ([]models.Book, int64, error) {
	var (
		articles []models.Book
		count    int64
	)
	repo.db.Model(&articles).Count(&count).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag asc")
		}).
		Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("created_at desc").Find(&articles)
	return articles, count, nil
}

func (repo *BookRepository) ListByTag(tag string, offset, limit int) ([]models.Book, int64, error) {
	var (
		t        models.Tag
		articles []models.Book
		count    int64
	)
	err := repo.db.Where(&models.Tag{Tag: tag}).First(&t).Error
	if err != nil {
		return nil, 0, err
	}
	repo.db.Model(&t).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag asc")
		}).
		Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("created_at desc").
		Association("Books").
		Find(&articles)
	count = repo.db.Model(&t).Association("Books").Count()
	return articles, count, nil
}

func (repo *BookRepository) ListByAuthor(username string, offset, limit int) ([]models.Book, int64, error) {
	var (
		u        models.User
		articles []models.Book
		count    int64
	)
	err := repo.db.Where(&models.User{Username: username}).First(&u).Error
	if err != nil {
		return nil, 0, err
	}
	repo.db.Where(&models.Book{UploadedByID: u.ID}).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag asc")
		}).
		Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("created_at desc").
		Find(&articles)
	repo.db.Where(&models.Book{UploadedByID: u.ID}).Model(&models.Book{}).Count(&count)
	return articles, count, nil
}

func (repo *BookRepository) ListFeed(userID uint, offset, limit int) ([]models.Book, int64, error) {
	var (
		u        models.User
		articles []models.Book
		count    int64
	)
	err := repo.db.First(&u, userID).Error
	if err != nil {
		return nil, 0, err
	}
	var followings []models.Follow

	repo.db.Model(&u).Preload("Following").Preload("Follower").Find(&followings)
	if len(followings) == 0 {
		return articles, 0, nil
	}
	ids := make([]uint, len(followings))
	for i, f := range followings {
		ids[i] = f.FollowingID
	}
	repo.db.Where("author_id in (?)", ids).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag asc")
		}).
		Preload("UploadedBy").
		Offset(offset).
		Limit(limit).
		Order("created_at desc").
		Find(&articles)
	repo.db.Where(&models.Book{UploadedByID: u.ID}).Model(&models.Book{}).Count(&count)

	return articles, count, nil
}

func (repo *BookRepository) AddComment(a *models.Book, c *models.Review) error {
	err := repo.db.Model(a).Association("Comments").Append(c)
	if err != nil {
		return err
	}

	return repo.db.Where(c.ID).Preload("User").First(c).Error
}

func (repo *BookRepository) GetCommentsBySlug(slug string) ([]models.Review, error) {
	var m models.Book
	err := repo.db.Where(&models.Book{Slug: slug}).Preload("Comments").Preload("Comments.User").First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.Reviews, nil
}

func (repo *BookRepository) GetCommentByID(id uint) (*models.Review, error) {
	var m models.Review
	if err := repo.db.Where(id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (repo *BookRepository) DeleteComment(c *models.Review) error {
	return repo.db.Delete(c).Error
}

func (repo *BookRepository) ListTags() ([]models.Tag, error) {
	var tags []models.Tag
	repo.db.Find(&tags)
	if len(tags) == 0 {
		return nil, errors.New("tags not found")
	}
	return tags, nil
}

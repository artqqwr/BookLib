package repository

import (
	"errors"

	"github.com/artqqwr/bookslib/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) GetByID(id uint) (*models.User, error) {
	var m models.User
	if err := repo.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (repo *UserRepository) GetByEmail(e string) (*models.User, error) {
	var m models.User
	if err := repo.db.Where(&models.User{Email: e}).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (repo *UserRepository) GetByUsername(username string) (*models.User, error) {
	var m models.User
	if err := repo.db.Where(&models.User{Username: username}).Preload("Followers").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (repo *UserRepository) Create(u *models.User) (err error) {
	return repo.db.Create(u).Error
}

func (repo *UserRepository) Update(u *models.User) error {
	return repo.db.Model(u).Updates(u).Error
}

func (repo *UserRepository) AddFollower(u *models.User, followerID uint) error {
	return repo.db.Model(u).Association("Followers").Append(&models.Follow{FollowerID: followerID, FollowingID: u.ID})
}

func (repo *UserRepository) RemoveFollower(u *models.User, followerID uint) error {
	if repo.db.Config.Dialector.Name() == "sqlite" {
		err := repo.db.Exec("delete from `follows` where `follower_id`=? and `following_id`=?",
			followerID,
			u.ID).Error

		if err != nil {
			return err
		}
		return nil
	} else {
		f := models.Follow{
			FollowerID:  followerID,
			FollowingID: u.ID,
		}
		if err := repo.db.Model(u).Association("Followers").Find(&f); err != nil {
			return err
		}
		if err := repo.db.Delete(f).Error; err != nil {
			return err
		}

		return nil
	}

}

func (repo *UserRepository) IsFollower(userID, followerID uint) (bool, error) {
	var f models.Follow
	if err := repo.db.Where("following_id = ? AND follower_id = ?", userID, followerID).First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

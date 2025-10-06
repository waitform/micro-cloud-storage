package model

import (
	"errors"

	"gorm.io/gorm"
)

type userDAOImpl struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &userDAOImpl{db: db}
}

func (d *userDAOImpl) CreateUser(user *User) error {
	return d.db.Create(user).Error
}

func (d *userDAOImpl) GetByID(id int64) (*User, error) {
	var user User
	err := d.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (d *userDAOImpl) GetByUsername(username string) (*User, error) {
	var user User
	err := d.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (d *userDAOImpl) UpdateUser(user *User) error {
	return d.db.Save(user).Error
}

func (d *userDAOImpl) UpdateUsage(userID int64, delta int64) error {
	return d.db.Model(&User{}).
		Where("id = ?", userID).
		UpdateColumn("used_space", gorm.Expr("used_space + ?", delta)).
		Error
}

func (d *userDAOImpl) UpdateCapacity(userID int64, newTotalSpace int64) error {
	return d.db.Model(&User{}).
		Where("id = ?", userID).
		Update("total_space", newTotalSpace).
		Error
}

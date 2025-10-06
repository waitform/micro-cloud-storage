package model

type UserDAO interface {
	CreateUser(user *User) error
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	UpdateUser(user *User) error
	UpdateUsage(userID int64, delta int64) error
	UpdateCapacity(userID int64, newTotalSpace int64) error
}

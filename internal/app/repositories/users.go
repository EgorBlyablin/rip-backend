package repositories

import (
	"rip/internal/app/ds"

	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UsersRepository struct{ usersDB *gorm.DB }

var (
	ErrorLoginIsTaken = errors.New("login is already taken")
	ErrorUserNotFound = errors.New("user not found")
)

func NewUsersRepository(db *gorm.DB) *UsersRepository {
	return &UsersRepository{
		usersDB: db,
	}
}

func (r *UsersRepository) GetUsers() ([]ds.User, error) {
	users := []ds.User{}

	err := r.usersDB.Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get users from DB")
		return []ds.User{}, err
	}

	return users, nil
}

func (r *UsersRepository) GetUserByID(userId uint) (ds.User, error) {
	user := ds.User{}

	err := r.usersDB.Where(&ds.User{ID: userId}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, ErrorUserNotFound
		}
		log.WithError(err).WithFields(log.Fields{
			"User ID": userId,
		}).Error("Failed to get user by ID from DB")
		return ds.User{}, err
	}

	return user, nil
}

func (r *UsersRepository) GetUserByLogin(login string) (ds.User, error) {
	user := ds.User{}

	err := r.usersDB.Where(&ds.User{Login: login}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, ErrorUserNotFound
		}
		return ds.User{}, err
	}

	return user, nil
}

func (r *UsersRepository) CreateUser(user ds.CreateUser) (ds.User, error) {
	var newUser = ds.User{
		Login:    user.Login,
		Password: user.Password,
	}

	if r.usersDB.First(&newUser).Error != nil {
		return ds.User{}, ErrorLoginIsTaken
	}

	err := r.usersDB.Create(&newUser).Error
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Login":    user.Login,
			"Password": "***",
		}).Error("Failed to create user in DB")
		return ds.User{}, err
	}

	return newUser, nil
}

func (r *UsersRepository) UpdateUser(userId uint, user ds.UpdateUser) (ds.User, error) {
	updatedUser := ds.User{}

	updates := map[string]any{}

	if user.Login != nil {
		updates["login"] = user.Login
	}
	if user.Password != nil {
		updates["password"] = user.Password
	}
	if len(updates) == 0 {
		return r.GetUserByID(userId)
	}

	err := r.usersDB.Model(&ds.User{}).Where(&ds.User{ID: userId}).Clauses(clause.Returning{}).Updates(updates).Scan(&updatedUser).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ds.User{}, ErrorLoginIsTaken
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"User ID":  userId,
			"Login":    user.Login,
			"Password": "***",
		}).Error("Failed to update user in DB")
		return ds.User{}, err
	}

	return updatedUser, nil
}

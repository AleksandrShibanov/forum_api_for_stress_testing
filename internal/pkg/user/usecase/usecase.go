package usecase

import (
	"forum/internal/models"
	"forum/internal/pkg/user/repository"

	myerror "forum/internal/error"
)

type UserUsecase struct {
	ur *repository.UserRepository
}

func NewUserUsecase(ur *repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		ur: ur,
	}
}

func (uu *UserUsecase) Create(user *models.User) (*models.User, error) {
	return uu.ur.Insert(user)
}

func (uu *UserUsecase) GetByNickname(nickname string) (*models.User, error) {
	return uu.ur.SelectByNickname(nickname)
}

func (uu *UserUsecase) GetByEmail(email string) (*models.User, error) {
	user, dbErr := uu.ur.SelectByEmail(email)

	switch dbErr {
	case myerror.DBScanError:
		return nil, myerror.UNotFound
	case nil:
		return user, nil
	}

	return nil, myerror.UnexpectedError
}

func (uu *UserUsecase) GetConflict(nickname string, email string) ([]*models.User, error) {
	return uu.ur.SelectConflict(nickname, email)
}

func (uu *UserUsecase) Update(nickname string, toUpdate *models.UserUpdate) (*models.User, error) {
	return uu.ur.Update(nickname, toUpdate)
}

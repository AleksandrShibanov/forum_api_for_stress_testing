package usecase

import (
	"forum/internal/models"
	"forum/internal/pkg/forum/repository"

	myerror "forum/internal/error"
)

type ForumUsecase struct {
	fr *repository.ForumRepository
}

func NewForumUsecase(fr *repository.ForumRepository) *ForumUsecase {
	return &ForumUsecase{
		fr: fr,
	}
}

func (fu *ForumUsecase) Create(forum *models.Forum) (*models.Forum, error) {
	dbErr := fu.fr.Insert(forum)

	switch dbErr {
	case myerror.DBCommitError, myerror.DBCreateTxError, myerror.DBRollbackError:
		return nil, myerror.UInternalError
	case myerror.DBScanError:
		return nil, myerror.UAlreadyExist
	case nil:
		return forum, nil
	}

	return nil, myerror.UnexpectedError
}

func (fu *ForumUsecase) GetBySlug(slug string) (*models.Forum, error) {
	forum, dbErr := fu.fr.SelectBySlug(slug)

	switch dbErr {
	case myerror.DBScanError:
		return nil, myerror.UNotFound
	case nil:
		return forum, nil
	}

	return nil, myerror.UnexpectedError
}

func (fu *ForumUsecase) GetUsersBySlug(slug string, since string, limit int64, isDescOrder bool) ([]*models.User, error) {
	return fu.fr.SelectUsersBySlug(slug, since, limit, isDescOrder)
}

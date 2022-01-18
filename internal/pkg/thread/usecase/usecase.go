package usecase

import (
	"forum/internal/models"
	"forum/internal/pkg/thread/repository"
	"strconv"
	"time"

	myerror "forum/internal/error"
)

type ThreadUsecase struct {
	tr *repository.ThreadRepository
}

func NewThreadUsecase(tr *repository.ThreadRepository) *ThreadUsecase {
	return &ThreadUsecase{
		tr: tr,
	}
}

func (tu *ThreadUsecase) Create(thread *models.Thread) (*models.Thread, error) {
	if thread.Created.IsZero() {
		thread.Created = time.Now()
	}

	return tu.tr.Insert(thread)
}

func (tu *ThreadUsecase) Get(id int32) (*models.Thread, error) {
	return tu.tr.Select(id)
}

func (tu *ThreadUsecase) GetBySlug(slug string) (*models.Thread, error) {
	return tu.tr.SelectBySlug(slug)
}

func (tu *ThreadUsecase) GetBySlugOrId(slug_or_id string) (*models.Thread, error) {
	var slug string
	var id int32
	passedId, passedSlug := false, false

	aId, err := strconv.Atoi(slug_or_id)
	if err != nil {
		slug = slug_or_id
		passedSlug = len(slug_or_id) != 0
	} else {
		id = int32(aId)
		passedId = true
	}

	var thread *models.Thread
	thread = nil

	if passedId {
		thread, err = tu.tr.Select(id)
		if err != nil {
			return nil, myerror.NotExist
		}
	} else if passedSlug {
		thread, err = tu.tr.SelectBySlug(slug)
		if err != nil {
			return nil, myerror.NotExist
		}
	}

	if thread == nil {
		return nil, myerror.InternalError
	}

	return thread, nil
}

func (tu *ThreadUsecase) UpdateBySlugOrId(slug_or_id string, threadToUpdate *models.ThreadUpdate) (*models.Thread, error) {
	var slug string
	var id int32
	passedId, passedSlug := false, false

	aId, err := strconv.Atoi(slug_or_id)
	if err != nil {
		slug = slug_or_id
		passedSlug = len(slug_or_id) != 0
	} else {
		id = int32(aId)
		passedId = true
	}

	var updatedThread *models.Thread
	updatedThread = nil

	if passedId {
		updatedThread, err = tu.tr.Update(int64(id), threadToUpdate)
		if err != nil {
			return nil, err
		}
	} else if passedSlug {
		updatedThread, err = tu.tr.UpdateBySlug(slug, threadToUpdate)
		if err != nil {
			return nil, err
		}
	}

	if updatedThread == nil {
		return nil, myerror.InternalError
	}

	return updatedThread, nil
}

func (tu *ThreadUsecase) GetAll(forum string, limit int64, since string, isDescOrder bool) ([]*models.Thread, error) {
	return tu.tr.SelectAll(forum, limit, since, isDescOrder)
}

package usecase

import (
	"forum/internal/models"
	"forum/internal/pkg/post/repository"
	threadRepository "forum/internal/pkg/thread/repository"
	"strconv"
	"time"

	myerror "forum/internal/error"
)

func Difference(a, b []int64) (diff []int64) {
	m := make(map[int64]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
			m[item] = true
		}
	}
	return
}

type PostUsecase struct {
	pr *repository.PostRepository
	tr *threadRepository.ThreadRepository
}

func NewPostUsecase(pr *repository.PostRepository, tr *threadRepository.ThreadRepository) *PostUsecase {
	return &PostUsecase{
		pr: pr,
		tr: tr,
	}
}

func (pu *PostUsecase) CreateAll(posts []*models.Post, slug_or_id string) ([]*models.Post, error) {
	insertTime := time.Now()
	var slug string
	var id int32
	passedId, passedSlug := false, false
	var pForum *string
	var pId *int64

	aId, err := strconv.Atoi(slug_or_id)
	if err != nil {
		slug = slug_or_id
		passedSlug = len(slug_or_id) != 0
		pId, pForum, err = pu.tr.SelectThreadIdForumBySlug(slug)
		if err != nil {
			return nil, myerror.NotExist
		}
	} else {
		id = int32(aId)
		passedId = true
		pForum, err = pu.tr.SelectForumByThreadId(int64(id))
		if err != nil {
			return nil, myerror.NotExist
		}
	}

	var parent_ids, ids []int64
	for _, post := range posts {

		if post.Created.IsZero() {
			post.Created = insertTime
		}

		if passedId {
			post.Thread = id
			post.Forum = *pForum
		} else if passedSlug {

			post.Thread = int32(*pId)
			post.Forum = *pForum
		}

		if post.Parent != 0 {
			parent_ids = append(parent_ids, post.Parent)
		}
		ids = append(ids, post.Id)
	}

	if len(posts) == 0 {
		return []*models.Post{}, nil
	}

	parentsToCheck := Difference(parent_ids, ids)
	if len(parentsToCheck) > 0 {
		noConflict, err := pu.CheckAllParentsExist(parentsToCheck, *pForum)
		if err != nil {
			return nil, myerror.InternalError
		} else if !noConflict {
			return nil, myerror.ConflictError
		}
	}

	return pu.pr.InsertAll(posts)
}

func (pu *PostUsecase) GetAll(slug_or_id string, limit int64, since int64, sort string, isDescOrder bool) ([]*models.Post, error) {
	var slug string
	var id int32

	aId, err := strconv.Atoi(slug_or_id)
	if err != nil {
		slug = slug_or_id
		pId, _, err := pu.tr.SelectThreadIdForumBySlug(slug)
		if err != nil {
			return nil, myerror.NotExist
		}
		id = int32(*pId)
	} else {
		id = int32(aId)
		_, err := pu.tr.SelectForumByThreadId(int64(id))
		if err != nil {
			return nil, myerror.NotExist
		}
	}

	if sort == "" || sort == "flat" {
		return pu.pr.SelectAllFlat(id, limit, since, isDescOrder)
	} else if sort == "tree" {
		return pu.pr.SelectAllTree(id, limit, since, isDescOrder)
	} else if sort == "parent_tree" {
		return pu.pr.SelectAllParentTree(id, limit, since, isDescOrder)
	}

	return nil, nil
}

func (pu *PostUsecase) Get(id int64) (*models.Post, error) {
	return pu.pr.Get(id)
}

func (pu *PostUsecase) CheckAllParentsExist(ids []int64, forum string) (bool, error) {
	return pu.pr.Check(ids, forum)
}

func (pu *PostUsecase) Update(id int64, postToUpdate *models.PostUpdate) (*models.Post, error) {
	return pu.pr.Update(int64(id), postToUpdate)
}

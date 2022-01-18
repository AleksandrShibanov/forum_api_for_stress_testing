package usecase

import (
	"forum/internal/models"
	threadRepository "forum/internal/pkg/thread/repository"
	"forum/internal/pkg/vote/repository"
	"strconv"

	myerror "forum/internal/error"
)

type VoteUsecase struct {
	vr *repository.VoteRepository
	tr *threadRepository.ThreadRepository
}

func NewVoteUsecase(vr *repository.VoteRepository, tr *threadRepository.ThreadRepository) *VoteUsecase {
	return &VoteUsecase{
		vr: vr,
		tr: tr,
	}
}

func (vu *VoteUsecase) Create(vote *models.Vote) (*models.Vote, error) {
	return vu.vr.Insert(vote)
}

func (vu *VoteUsecase) CreateBySlugOrId(vote *models.Vote, slug_or_id string) (*models.Thread, error) {
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

	if passedId {
		vote.Thread = id
	} else if passedSlug {
		pId, _, err := vu.tr.SelectThreadIdForumBySlug(slug)
		if err != nil {
			return nil, myerror.NotExist
		}
		vote.Thread = int32(*pId)
	}

	newVote, createErr := vu.Create(vote)
	if createErr != nil {
		return nil, createErr
	}

	thread, updateErr := vu.tr.Select(newVote.Thread)
	if updateErr != nil {
		return nil, updateErr
	}

	return thread, nil
}

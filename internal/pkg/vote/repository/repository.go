package repository

import (
	"context"
	"database/sql"
	myerror "forum/internal/error"
	"forum/internal/models"
)

type VoteRepository struct {
	DB *sql.DB
}

func NewVoteRepository(DB *sql.DB) *VoteRepository {
	return &VoteRepository{
		DB: DB,
	}
}

func (pr *VoteRepository) Insert(vote *models.Vote) (*models.Vote, error) {
	tx, err := pr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	newVote := &models.Vote{}
	err = tx.QueryRow(`INSERT INTO vote (author, voice, thread)
	VALUES ($1, $2, $3) ON CONFLICT (author, thread)
	DO
	UPDATE
	SET voice=$2 RETURNING author, voice, thread`, vote.Nickname, vote.Voice, vote.Thread).Scan(&newVote.Nickname, &newVote.Voice, &newVote.Thread)

	if err != nil {
		tx.Rollback()
		return nil, myerror.NotExist
	}

	err = tx.Commit()
	if err != nil {
		return nil, myerror.NotExist
	}

	return newVote, nil
}

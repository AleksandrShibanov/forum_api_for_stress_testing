package repository

import (
	"context"
	"database/sql"
	"fmt"
	myerror "forum/internal/error"
	"forum/internal/models"
	"regexp"
)

type ThreadRepository struct {
	DB *sql.DB
}

func NewThreadRepository(DB *sql.DB) *ThreadRepository {
	return &ThreadRepository{
		DB: DB,
	}
}

func (tr *ThreadRepository) Insert(thread *models.Thread) (*models.Thread, error) {
	tx, err := tr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	newThread := &models.Thread{}
	err = tx.QueryRow(`INSERT INTO thread (title, author, forum, message, votes, slug, created_at)
	VALUES ($1, $2, COALESCE((SELECT slug FROM forum WHERE slug = $3), $3), $4, $5, $6, $7) RETURNING id, title, author, forum, message, votes, slug, created_at;`,
		thread.Title, thread.Author, thread.Forum, thread.Message, thread.Votes, thread.Slug, thread.Created).Scan(&newThread.Id, &newThread.Title, &newThread.Author, &newThread.Forum, &newThread.Message, &newThread.Votes, &newThread.Slug, &newThread.Created)

	if err != nil {
		tx.Rollback()
		if match, _ := regexp.MatchString(`.*violates foreign.*`, err.Error()); match {
			return nil, myerror.NotExist
		} else {
			return nil, myerror.ConflictError
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, myerror.InternalError
	}

	return newThread, nil
}

func (tr *ThreadRepository) Select(id int32) (*models.Thread, error) {
	thread := models.Thread{}
	var buf sql.NullString

	err := tr.DB.QueryRow(
		"SELECT id, title, author, forum, message, votes, slug, created_at FROM thread WHERE id = $1", id).
		Scan(&thread.Id, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &buf, &thread.Created)

	if err != nil {
		return nil, myerror.InternalError
	}

	if buf.Valid {
		thread.Slug = buf.String
	}

	return &thread, nil
}

func (tr *ThreadRepository) SelectBySlug(slug string) (*models.Thread, error) {
	row := tr.DB.QueryRow(
		"SELECT id, title, author, forum, message, votes, slug, created_at FROM thread WHERE slug = $1",
		slug)

	thread := models.Thread{}
	var buf sql.NullString
	err := row.Scan(&thread.Id, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &buf, &thread.Created)
	if err != nil {
		return nil, myerror.InternalError
	}

	if buf.Valid {
		thread.Slug = buf.String
	}

	return &thread, nil
}

func (tr *ThreadRepository) SelectAll(forum string, limit int64, since string, isDescOrder bool) ([]*models.Thread, error) {
	var exists bool
	err := tr.DB.QueryRow("SELECT exists (SELECT id FROM thread WHERE forum=$1)", forum).Scan(&exists)
	if err != nil || !exists {
		return nil, myerror.NotExist
	}

	query := "SELECT id, title, author, forum, message, votes, slug, created_at FROM thread WHERE forum = $1"
	arr := []interface{}{
		forum,
	}

	if len(since) > 0 {
		query += " AND created_at"
		if isDescOrder {
			query += " <="
		} else {
			query += " >="
		}
		query += fmt.Sprintf(" $%d", len(arr)+1)
		arr = append(arr, since)
	}

	query += " ORDER BY created_at"

	if isDescOrder {
		query += " DESC"
	} else {
		query += " ASC"
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(arr)+1)
		arr = append(arr, limit)
	}

	rows, err := tr.DB.Query(query, arr...)
	if err != nil {
		return nil, myerror.InternalError
	}
	defer rows.Close()

	threads := []*models.Thread{}

	for rows.Next() {
		thread := models.Thread{}
		var buf sql.NullString
		if err := rows.Scan(&thread.Id, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &buf, &thread.Created); err != nil {
			return nil, myerror.InternalError
		}
		if buf.Valid {
			thread.Slug = buf.String
		}

		threads = append(threads, &thread)
	}

	return threads, nil
}

func (tr *ThreadRepository) SelectForumByThreadId(id int64) (*string, error) {
	row := tr.DB.QueryRow("SELECT forum FROM thread WHERE id = $1", id)

	var forum string
	err := row.Scan(&forum)
	if err != nil {
		return nil, myerror.InternalError
	}

	return &forum, nil
}

func (tr *ThreadRepository) SelectThreadIdForumBySlug(slug string) (*int64, *string, error) {
	row := tr.DB.QueryRow("SELECT id, forum FROM thread WHERE slug = $1", slug)

	var id int64
	var forum string
	err := row.Scan(&id, &forum)
	if err != nil {
		return nil, nil, myerror.InternalError
	}

	return &id, &forum, nil
}

func (tr *ThreadRepository) Update(id int64, threadToUpdate *models.ThreadUpdate) (*models.Thread, error) {
	tx, err := tr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	thread := &models.Thread{}
	err = tx.QueryRow(`UPDATE thread SET message = COALESCE($2, message), title = COALESCE($3, title) WHERE id = $1
	RETURNING id, author, title, forum, message, votes, slug, created_at`, id, threadToUpdate.Message, threadToUpdate.Title).
		Scan(&thread.Id, &thread.Author, &thread.Title, &thread.Forum, &thread.Message, &thread.Votes, &thread.Slug, &thread.Created)

	if err != nil {
		tx.Rollback()
		if match, _ := regexp.MatchString(`.*no rows.*`, err.Error()); match {
			return nil, myerror.NotExist
		} else {
			return nil, myerror.ConflictError
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, myerror.InternalError
	}

	return thread, nil
}

func (tr *ThreadRepository) UpdateBySlug(slug string, threadToUpdate *models.ThreadUpdate) (*models.Thread, error) {
	tx, err := tr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	thread := &models.Thread{}
	err = tx.QueryRow(`UPDATE thread SET message = COALESCE($2, message), title = COALESCE($3, title) WHERE slug = $1
	RETURNING id, author, title, forum, message, votes, slug, created_at`, slug, threadToUpdate.Message, threadToUpdate.Title).
		Scan(&thread.Id, &thread.Author, &thread.Title, &thread.Forum, &thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
	if err != nil {
		tx.Rollback()
		if match, _ := regexp.MatchString(`.*no rows.*`, err.Error()); match {
			return nil, myerror.NotExist
		} else {
			return nil, myerror.ConflictError
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, myerror.InternalError
	}

	return thread, nil
}

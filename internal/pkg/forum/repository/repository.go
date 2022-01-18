package repository

import (
	"context"
	"database/sql"
	"fmt"
	myerror "forum/internal/error"
	"forum/internal/models"
)

type ForumRepository struct {
	DB *sql.DB
}

func NewForumRepository(DB *sql.DB) *ForumRepository {
	return &ForumRepository{
		DB: DB,
	}
}

func (fr *ForumRepository) Insert(forum *models.Forum) error {
	tx, err := fr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return myerror.DBCreateTxError
	}

	err = tx.QueryRow(`INSERT INTO forum (title, author, slug, posts, threads)
						VALUES ($1, COALESCE((SELECT nickname FROM users WHERE nickname = $2), $2), $3, $4, $5) RETURNING title, author, slug, posts, threads;`,
		forum.Title, forum.User, forum.Slug, forum.Posts, forum.Threads).Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)

	if err != nil {
		tx.Rollback()
		return myerror.DBScanError
	}

	err = tx.Commit()
	if err != nil {
		return myerror.DBCommitError
	}

	return nil
}

func (fr *ForumRepository) SelectBySlug(slug string) (*models.Forum, error) {
	row := fr.DB.QueryRow(
		"SELECT title, author, slug, posts, threads FROM forum WHERE slug = $1",
		slug)

	forum := models.Forum{}
	buf := sql.NullString{}
	err := row.Scan(&forum.Title, &buf, &forum.Slug, &forum.Posts, &forum.Threads)
	if err != nil {
		return nil, myerror.DBScanError
	}

	if buf.Valid {
		forum.User = buf.String
	}

	return &forum, nil
}

func (fr *ForumRepository) SelectUsersBySlug(slug string, since string, limit int64, isDescOrder bool) ([]*models.User, error) {
	var exists bool
	err := fr.DB.QueryRow("SELECT exists (SELECT id FROM forum WHERE slug=$1)", slug).Scan(&exists)
	if err != nil || !exists {
		return nil, myerror.NotExist
	}

	query := `SELECT nickname, fullname, about, email FROM forum_users where forum=$1`
	arr := []interface{}{
		slug,
	}

	var order string
	var sign string
	if isDescOrder {
		order = "DESC"
		sign = "<"
	} else {
		order = "ASC"
		sign = ">"
	}

	if len(since) > 0 {
		query += fmt.Sprintf(" AND nickname %s $%d", sign, len(arr)+1)
		arr = append(arr, since)
	}

	query += " ORDER BY nickname "
	query += order

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(arr)+1)
		arr = append(arr, limit)
	}

	rows, err := fr.DB.Query(query, arr...)
	if err != nil {
		return nil, myerror.NotExist
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {

		user := models.User{}
		if err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email); err != nil {
			return nil, myerror.InternalError
		}
		users = append(users, &user)
	}
	return users, nil
}

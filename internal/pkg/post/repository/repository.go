package repository

import (
	"context"
	"database/sql"
	"fmt"
	myerror "forum/internal/error"
	"forum/internal/models"
)

type PostRepository struct {
	DB *sql.DB
}

func NewPostRepository(DB *sql.DB) *PostRepository {
	return &PostRepository{
		DB: DB,
	}
}

func (pr *PostRepository) InsertAll(posts []*models.Post) ([]*models.Post, error) {
	tx, err := pr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	query := `INSERT INTO post (parent, author, message, is_edited, forum, thread, created_at)
	VALUES`
	arr := []interface{}{}
	first := true
	for _, post := range posts {
		if !first {
			query += ","
		}
		query += fmt.Sprintf(" ($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(arr)+1, len(arr)+2, len(arr)+3, len(arr)+4, len(arr)+5, len(arr)+6, len(arr)+7)
		arr = append(arr, post.Parent)
		arr = append(arr, post.Author)
		arr = append(arr, post.Message)
		arr = append(arr, post.IsEdited)
		arr = append(arr, post.Forum)
		arr = append(arr, post.Thread)
		arr = append(arr, post.Created)
		first = false
	}
	query += " RETURNING id, parent, author, message, is_edited, forum, thread, created_at;"
	rows, err := tx.Query(query, arr...)
	defer rows.Close()

	if err != nil {
		tx.Rollback()
		return nil, myerror.ConflictError
	}

	i := 0
	newPosts := []*models.Post{}
	for rows.Next() {
		newPost := models.Post{}
		if err := rows.Scan(&newPost.Id, &newPost.Parent, &newPost.Author, &newPost.Message, &newPost.IsEdited, &newPost.Forum, &newPost.Thread, &newPost.Created); err != nil {
			tx.Rollback()
			return nil, myerror.InsertError
		}
		i++
		newPosts = append(newPosts, &newPost)
	}

	err = tx.Commit()
	if err != nil {
		return nil, myerror.NotExist
	}

	if len(newPosts) != len(posts) {
		return nil, myerror.ConflictError
	}

	return newPosts, nil
}

func (fr *PostRepository) SelectAllFlat(thread int32, limit int64, since int64, isDescOrder bool) ([]*models.Post, error) {
	query := "SELECT id, parent, author, message, is_edited, forum, thread, created_at FROM post WHERE thread = $1"
	arr := []interface{}{
		thread,
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

	if since > 0 {
		query += fmt.Sprintf(" AND id %s $%d", sign, len(arr)+1)
		arr = append(arr, since)
	}

	query += " ORDER BY created_at"
	query += " " + order

	query += ", id"
	query += " " + order

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(arr)+1)
		arr = append(arr, limit)
	}

	rows, _ := fr.DB.Query(query, arr...)
	defer rows.Close()

	posts := []*models.Post{}
	for rows.Next() {

		post := models.Post{}
		if err := rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created); err != nil {
			return nil, myerror.InternalError
		}
		posts = append(posts, &post)
	}

	return posts, nil
}

func (fr *PostRepository) SelectAllTree(thread int32, limit int64, since int64, isDescOrder bool) ([]*models.Post, error) {
	query := "SELECT id, parent, author, message, is_edited, forum, thread, created_at FROM post WHERE thread = $1"
	arr := []interface{}{
		thread,
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

	if since > 0 {
		query += fmt.Sprintf(" AND array_append(path, id) %s (SELECT array_append(path, id) FROM post WHERE id = $%d)", sign, len(arr)+1)
		arr = append(arr, since)
	}

	query += " ORDER BY array_append(path, id)"
	query += " " + order

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(arr)+1)
		arr = append(arr, limit)
	}

	rows, _ := fr.DB.Query(query, arr...)
	defer rows.Close()

	posts := []*models.Post{}
	for rows.Next() {
		post := models.Post{}
		if err := rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created); err != nil {
			return nil, myerror.InternalError
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (fr *PostRepository) SelectAllParentTree(thread int32, limit int64, since int64, isDescOrder bool) ([]*models.Post, error) {
	query := "SELECT id, parent, author, message, is_edited, forum, thread, created_at FROM post AS temp WHERE thread = $1"
	arr := []interface{}{
		thread,
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

	if limit > 0 {
		query += fmt.Sprintf(" AND (array_append(path, id))[1] IN (SELECT id FROM post WHERE thread = $%d AND parent=0", len(arr)+1)
		arr = append(arr, thread)
		if since > 0 {
			query += fmt.Sprintf(" AND (array_append(path, id))[1] %s (SELECT (array_append(path, id))[1] FROM post WHERE id = $%d)", sign, len(arr)+1)
			arr = append(arr, since)
		}
		query += fmt.Sprintf(" ORDER BY id %s LIMIT $%d)", order, len(arr)+1)
		arr = append(arr, limit)
	}

	query += " ORDER BY (array_append(path, id))[1]"
	query += " " + order

	query += ", (array_append(path, id))[2:]"

	rows, _ := fr.DB.Query(query, arr...)
	defer rows.Close()

	posts := []*models.Post{}
	for rows.Next() {

		post := models.Post{}
		if err := rows.Scan(&post.Id, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created); err != nil {
			return nil, myerror.InternalError
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (pr *PostRepository) Get(id int64) (*models.Post, error) {
	row := pr.DB.QueryRow(`SELECT parent, author, message, is_edited, forum, thread, created_at FROM post WHERE id = $1`, id)

	post := models.Post{
		Id: id,
	}
	err := row.Scan(&post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created)
	if err != nil {
		return nil, myerror.NotExist
	}

	return &post, nil
}

func (pr *PostRepository) Check(ids []int64, forum string) (bool, error) {
	query := fmt.Sprintf("select %d = (select count(*) from post where forum=$1 and id in (", len(ids))
	isFirst := true
	for _, id := range ids {
		if isFirst {
			query += fmt.Sprintf("'%d'", id)
			isFirst = false
		} else {
			query += fmt.Sprintf(", '%d'", id)
		}
	}
	query += "))"
	row := pr.DB.QueryRow(query, forum)

	noConflict := false
	err := row.Scan(&noConflict)
	if err != nil {
		return false, myerror.InternalError
	}

	return noConflict, nil
}

func (pr *PostRepository) Update(id int64, postToUpdate *models.PostUpdate) (*models.Post, error) {
	tx, err := pr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	newPost := &models.Post{}
	err = tx.QueryRow(`UPDATE post SET message = COALESCE($2, message), is_edited=(CASE WHEN $2 IS NULL OR message=$2 THEN is_edited ELSE true END) WHERE id = $1
	RETURNING id, parent, author, message, is_edited, forum, thread, created_at`, id, postToUpdate.Message).
		Scan(&newPost.Id, &newPost.Parent, &newPost.Author, &newPost.Message, &newPost.IsEdited, &newPost.Forum, &newPost.Thread, &newPost.Created)
	if err != nil {
		tx.Rollback()
		return nil, myerror.BadUpdate
	}

	err = tx.Commit()
	if err != nil {
		return nil, myerror.ConflictError
	}

	return newPost, nil
}

package repository

import (
	"database/sql"
	"forum/internal/models"
)

type ServiceRepository struct {
	DB *sql.DB
}

func NewServiceRepository(DB *sql.DB) *ServiceRepository {
	return &ServiceRepository{
		DB: DB,
	}
}

func (sr *ServiceRepository) CountAll() (*models.Status, error) {
	query := `
	SELECT * FROM
	(SELECT COUNT(*) FROM users) AS u,
	(SELECT COUNT(*) FROM forum) AS f,
	(SELECT COUNT(*) FROM thread) AS t,
	(SELECT COUNT(*) FROM post) AS p`

	status := &models.Status{}
	err := sr.DB.QueryRow(query).Scan(&status.User, &status.Forum, &status.Thread, &status.Post)

	return status, err
}

func (sr *ServiceRepository) Clear() error {
	query := `TRUNCATE users, forum, thread, post, vote, forum_users`
	_, err := sr.DB.Exec(query)

	return err
}

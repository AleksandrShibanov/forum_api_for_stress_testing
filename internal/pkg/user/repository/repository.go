package repository

import (
	"context"
	"database/sql"
	myerror "forum/internal/error"
	"forum/internal/models"
	"regexp"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(DB *sql.DB) *UserRepository {
	return &UserRepository{
		DB: DB,
	}
}

func (ur *UserRepository) Insert(user *models.User) (*models.User, error) {
	tx, err := ur.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	newUser := &models.User{}
	err = tx.QueryRow(`INSERT INTO users (nickname, fullname, about, email)
						VALUES ($1, $2, $3, $4) RETURNING nickname, fullname, about, email;`,
		user.Nickname, user.Fullname, user.About, user.Email).Scan(&newUser.Nickname, &newUser.Fullname, &newUser.About, &newUser.Email)

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

	return newUser, nil
}

func (ur *UserRepository) SelectByNickname(nickname string) (*models.User, error) {
	user := &models.User{}

	err := ur.DB.QueryRow(
		"SELECT nickname, fullname, about, email FROM users WHERE nickname = $1",
		nickname).Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)

	if err != nil {
		if match, _ := regexp.MatchString(`.*no rows.*`, err.Error()); match {
			return nil, myerror.NotExist
		} else {
			return nil, myerror.ConflictError
		}
	}

	return user, nil
}

func (ur *UserRepository) SelectByEmail(email string) (*models.User, error) {
	row := ur.DB.QueryRowContext(context.Background(),
		"SELECT nickname, fullname, about, email FROM users WHERE email = $1",
		email)

	user := models.User{}
	err := row.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
	if err != nil {
		return nil, myerror.DBScanError
	}

	return &user, nil
}

func (ur *UserRepository) SelectConflict(nickname string, email string) ([]*models.User, error) {
	rows, err := ur.DB.Query(
		"SELECT nickname, fullname, about, email FROM users WHERE nickname = $1 OR email = $2", nickname, email)
	if err != nil {
		return nil, myerror.DBSelectError
	}
	defer rows.Close()

	users := []*models.User{}

	for rows.Next() {
		user := models.User{}
		if err := rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email); err != nil {
			return nil, myerror.DBScanError
		}

		users = append(users, &user)
	}

	return users, nil
}

func (ur *UserRepository) Update(nickname string, toUpdate *models.UserUpdate) (*models.User, error) {
	tx, err := ur.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, myerror.InternalError
	}

	user := &models.User{}
	err = tx.QueryRow(`UPDATE users set fullname = COALESCE($2, fullname), about = COALESCE($3, about), email = COALESCE($4, email) WHERE nickname = $1 RETURNING nickname, fullname, about, email`, nickname, toUpdate.Fullname, toUpdate.About, toUpdate.Email).
		Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
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
		return nil, myerror.ConflictError
	}

	return user, nil
}

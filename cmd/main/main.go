package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"net/http"

	_ "github.com/jackc/pgx/stdlib"

	forumHandle "forum/internal/pkg/forum/delivery"
	forumRepo "forum/internal/pkg/forum/repository"
	forumUse "forum/internal/pkg/forum/usecase"

	userHandle "forum/internal/pkg/user/delivery"
	userRepo "forum/internal/pkg/user/repository"
	userUse "forum/internal/pkg/user/usecase"

	threadHandle "forum/internal/pkg/thread/delivery"
	threadRepo "forum/internal/pkg/thread/repository"
	threadUse "forum/internal/pkg/thread/usecase"

	postHandle "forum/internal/pkg/post/delivery"
	postRepo "forum/internal/pkg/post/repository"
	postUse "forum/internal/pkg/post/usecase"

	voteHandle "forum/internal/pkg/vote/delivery"
	voteRepo "forum/internal/pkg/vote/repository"
	voteUse "forum/internal/pkg/vote/usecase"

	serviceHandle "forum/internal/pkg/service/delivery"
	serviceRepo "forum/internal/pkg/service/repository"
	serviceUse "forum/internal/pkg/service/usecase"

	"github.com/gorilla/mux"
)

func getPostgres(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalln("cant open pgx", err)
	}
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		log.Fatalln(err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(time.Minute * 3)
	return db
}

func main() {

	sqlDB := getPostgres("postgres://aleksandr:password@127.0.0.1:5432/forum")
	defer sqlDB.Close()

	r := mux.NewRouter()
	r = r.PathPrefix("/api").Subrouter()

	fr := forumRepo.NewForumRepository(sqlDB)
	fu := forumUse.NewForumUsecase(fr)
	fh := forumHandle.NewUserHandler(fu)
	fh.Routing(r)

	ur := userRepo.NewUserRepository(sqlDB)
	uu := userUse.NewUserUsecase(ur)
	uh := userHandle.NewUserHandler(uu)
	uh.Routing(r)

	tr := threadRepo.NewThreadRepository(sqlDB)
	tu := threadUse.NewThreadUsecase(tr)
	th := threadHandle.NewThreadHandler(tu)
	th.Routing(r)

	pr := postRepo.NewPostRepository(sqlDB)
	pu := postUse.NewPostUsecase(pr, tr)
	ph := postHandle.NewPostHandler(pu, uu, tu, fu)
	ph.Routing(r)

	vr := voteRepo.NewVoteRepository(sqlDB)
	vu := voteUse.NewVoteUsecase(vr, tr)
	vh := voteHandle.NewVoteHandler(vu, tu)
	vh.Routing(r)

	sr := serviceRepo.NewServiceRepository(sqlDB)
	su := serviceUse.NewServiceUsecase(sr)
	sh := serviceHandle.NewServiceHandler(su)
	sh.Routing(r)

	fmt.Printf("start serving ::%s\n", "5000")

	error := http.ListenAndServe(":5000", r)
	log.Fatalf("http serve error %v", error)
}

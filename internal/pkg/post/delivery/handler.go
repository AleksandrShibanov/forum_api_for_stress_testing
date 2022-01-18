package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"forum/internal/models"
	forum "forum/internal/pkg/forum/usecase"
	"forum/internal/pkg/post/usecase"
	thread "forum/internal/pkg/thread/usecase"
	user "forum/internal/pkg/user/usecase"

	"github.com/gorilla/mux"

	myerror "forum/internal/error"
)

type PostHandler struct {
	pu *usecase.PostUsecase
	uu *user.UserUsecase
	tu *thread.ThreadUsecase
	fu *forum.ForumUsecase
}

func NewPostHandler(pu *usecase.PostUsecase, uu *user.UserUsecase, tu *thread.ThreadUsecase, fu *forum.ForumUsecase) *PostHandler {
	return &PostHandler{
		pu: pu,
		uu: uu,
		tu: tu,
		fu: fu,
	}
}

func (ph *PostHandler) Routing(r *mux.Router) {
	r.HandleFunc("/thread/{slug_or_id}/create", http.HandlerFunc(ph.CreatePosts)).Methods(http.MethodPost)
	r.HandleFunc(`/thread/{slug_or_id}/posts`, http.HandlerFunc(ph.GetAllPostsInThread)).Methods(http.MethodGet)
	r.HandleFunc(`/post/{id}/details`, http.HandlerFunc(ph.GetPost)).Methods(http.MethodGet)
	r.HandleFunc(`/post/{id}/details`, http.HandlerFunc(ph.UpdatePost)).Methods(http.MethodPost)
}

func (ph *PostHandler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	posts := []*models.Post{}

	err := json.NewDecoder(r.Body).Decode(&posts)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	slug_or_id := mux.Vars(r)["slug_or_id"]

	createdThread, createErr := ph.pu.CreateAll(posts, slug_or_id)
	if createErr == myerror.ConflictError {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(createErr)
		return
	}
	if createErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(createErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdThread)
}

func (ph *PostHandler) GetAllPostsInThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	slug_or_id := mux.Vars(r)["slug_or_id"]

	u, _ := url.Parse(r.URL.RequestURI())
	query := u.Query()

	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	since, _ := strconv.ParseInt(query.Get("since"), 10, 64)
	sort := query.Get("sort")
	desc_enabled := query.Get("desc")
	isDescOrder := false
	if desc_enabled == "true" {
		isDescOrder = true
	}

	selectedPosts, selectErr := ph.pu.GetAll(slug_or_id, limit, since, sort, isDescOrder)
	if selectErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(selectErr)
		return
	}
	if selectErr == myerror.ConflictError {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(selectErr)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(selectedPosts)
}

func (ph *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	postToUpdate := &models.PostUpdate{}

	err := json.NewDecoder(r.Body).Decode(postToUpdate)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	updatedPost, updateErr := ph.pu.Update(id, postToUpdate)
	if updateErr != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(postToUpdate)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPost)
}

func (ph *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	u, _ := url.Parse(r.URL.RequestURI())
	query := u.Query()
	related := query.Get("related")

	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var user *models.User
	user = nil
	var thread *models.Thread
	thread = nil
	var forum *models.Forum
	forum = nil

	post, selectErr := ph.pu.Get(id)
	if selectErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(selectErr)
		return
	}

	if related == "user" {
		user, _ = ph.uu.GetByNickname(post.Author)
	}

	if related == "thread" {
		thread, _ = ph.tu.GetBySlugOrId(fmt.Sprintf("%d", post.Thread))
	}

	if related == "forum" {
		forum, _ = ph.fu.GetBySlug(post.Forum)
	}

	if related == "user,thread" {
		user, _ = ph.uu.GetByNickname(post.Author)
		thread, _ = ph.tu.GetBySlugOrId(fmt.Sprintf("%d", post.Thread))
	}

	if related == "thread,forum" {
		thread, _ = ph.tu.GetBySlugOrId(fmt.Sprintf("%d", post.Thread))
		forum, _ = ph.fu.GetBySlug(post.Forum)
	}

	if related == "user,forum" {
		user, _ = ph.uu.GetByNickname(post.Author)
		forum, _ = ph.fu.GetBySlug(post.Forum)
	}

	if related == "user,thread,forum" {
		user, _ = ph.uu.GetByNickname(post.Author)
		thread, _ = ph.tu.GetBySlugOrId(fmt.Sprintf("%d", post.Thread))
		forum, _ = ph.fu.GetBySlug(post.Forum)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.PostFull{Post: post, Author: user, Thread: thread, Forum: forum})
}

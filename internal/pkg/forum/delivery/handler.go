package delivery

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"forum/internal/models"
	"forum/internal/pkg/forum/usecase"

	"github.com/gorilla/mux"

	myerror "forum/internal/error"
)

type ForumHandler struct {
	fu *usecase.ForumUsecase
}

func NewUserHandler(fu *usecase.ForumUsecase) *ForumHandler {
	return &ForumHandler{
		fu: fu,
	}
}

func (fh *ForumHandler) Routing(r *mux.Router) {
	r.HandleFunc("/forum/create", http.HandlerFunc(fh.CreateForum)).Methods(http.MethodPost)
	r.HandleFunc(`/forum/{slug}/details`, http.HandlerFunc(fh.ForumDetails)).Methods(http.MethodGet)
	r.HandleFunc(`/forum/{slug}/users`, http.HandlerFunc(fh.GetUsers)).Methods(http.MethodGet)
}

func (fh *ForumHandler) CreateForum(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	forum := &models.Forum{}

	err := json.NewDecoder(r.Body).Decode(forum)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	createdForum, createErr := fh.fu.Create(forum)
	if createErr == myerror.UAlreadyExist {
		selectedForum, selectErr := fh.fu.GetBySlug(forum.Slug)
		if selectErr != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(selectErr)
			return
		}

		if selectedForum.User == "" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(myerror.UNotFound)
			return
		}

		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(selectedForum)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdForum)
}

func (fh *ForumHandler) ForumDetails(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]

	forum, err := fh.fu.GetBySlug(slug)
	if err != nil || forum.User == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(myerror.UNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(forum)
}

func (fh *ForumHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]

	u, _ := url.Parse(r.URL.RequestURI())
	query := u.Query()

	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	since := query.Get("since")
	desc_enabled := query.Get("desc")
	isDescOrder := false
	if desc_enabled == "true" {
		isDescOrder = true
	}

	users, err := fh.fu.GetUsersBySlug(slug, since, limit, isDescOrder)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(myerror.UNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

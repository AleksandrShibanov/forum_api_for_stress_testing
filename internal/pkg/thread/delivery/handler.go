package delivery

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"forum/internal/models"
	"forum/internal/pkg/thread/usecase"

	"github.com/gorilla/mux"

	myerror "forum/internal/error"
)

type ThreadHandler struct {
	tu *usecase.ThreadUsecase
}

func NewThreadHandler(tu *usecase.ThreadUsecase) *ThreadHandler {
	return &ThreadHandler{
		tu: tu,
	}
}

func (th *ThreadHandler) Routing(r *mux.Router) {
	r.HandleFunc("/forum/{slug}/create", http.HandlerFunc(th.CreateThread)).Methods(http.MethodPost)
	r.HandleFunc(`/forum/{slug}/threads`, http.HandlerFunc(th.GetAllThreadsInForum)).Methods(http.MethodGet)
	r.HandleFunc(`/thread/{slug_or_id}/details`, http.HandlerFunc(th.GetThread)).Methods(http.MethodGet)
	r.HandleFunc(`/thread/{slug_or_id}/details`, http.HandlerFunc(th.UpdateThread)).Methods(http.MethodPost)
}

func (th *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	thread := &models.Thread{}

	err := json.NewDecoder(r.Body).Decode(thread)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	thread.Forum = mux.Vars(r)["slug"]

	createdThread, createErr := th.tu.Create(thread)
	if createErr == myerror.ConflictError {
		selectedThread, _ := th.tu.GetBySlug(thread.Slug)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(selectedThread)
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

func (th *ThreadHandler) GetAllThreadsInForum(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	slug := mux.Vars(r)["slug"]

	u, _ := url.Parse(r.URL.RequestURI())
	query := u.Query()

	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	since := query.Get("since")
	desc_enabled := query.Get("desc")
	isDescOrder := false
	if desc_enabled == "true" {
		isDescOrder = true
	}

	selectedThreads, selectErr := th.tu.GetAll(slug, limit, since, isDescOrder)
	if selectErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(selectErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(selectedThreads)
}

func (th *ThreadHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	slug_or_id := mux.Vars(r)["slug_or_id"]

	selectedThread, selectErr := th.tu.GetBySlugOrId(slug_or_id)
	if selectErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(selectErr)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(selectedThread)
}

func (th *ThreadHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	threadToUpdate := &models.ThreadUpdate{}

	err := json.NewDecoder(r.Body).Decode(threadToUpdate)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	slug_or_id := mux.Vars(r)["slug_or_id"]

	updatedThread, updateErr := th.tu.UpdateBySlugOrId(slug_or_id, threadToUpdate)
	if updateErr != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(updateErr)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedThread)
}

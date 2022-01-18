package delivery

import (
	"encoding/json"
	"net/http"

	"forum/internal/models"
	"forum/internal/pkg/user/usecase"

	"github.com/gorilla/mux"

	myerror "forum/internal/error"
)

type UserHandler struct {
	uu *usecase.UserUsecase
}

func NewUserHandler(uu *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		uu: uu,
	}
}

func (uh *UserHandler) Routing(r *mux.Router) {
	s := r.PathPrefix("/user").Subrouter()
	s.HandleFunc("/{nickname}/create", http.HandlerFunc(uh.Create)).Methods(http.MethodPost)
	s.HandleFunc("/{nickname}/profile", http.HandlerFunc(uh.Profile)).Methods(http.MethodGet)
	s.HandleFunc("/{nickname}/profile", http.HandlerFunc(uh.Update)).Methods(http.MethodPost)
}

func (uh *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	user := &models.User{
		Nickname: nickname,
	}

	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	createdUser, createErr := uh.uu.Create(user)
	if createErr == myerror.ConflictError {
		conflictUsers, _ := uh.uu.GetConflict(user.Nickname, user.Email)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(conflictUsers)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (uh *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nickname := vars["nickname"]

	user, err := uh.uu.GetByNickname(nickname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(myerror.UNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (uh *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	toUpdate := &models.UserUpdate{}

	err := json.NewDecoder(r.Body).Decode(toUpdate)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	updatedUser, updateErr := uh.uu.Update(nickname, toUpdate)
	if updateErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(updateErr)
		return
	}
	if updateErr == myerror.ConflictError {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(updateErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

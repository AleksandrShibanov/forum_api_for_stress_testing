package delivery

import (
	"encoding/json"
	"net/http"

	"forum/internal/models"
	thread "forum/internal/pkg/thread/usecase"
	"forum/internal/pkg/vote/usecase"

	myerror "forum/internal/error"

	"github.com/gorilla/mux"
)

type VoteHandler struct {
	vu *usecase.VoteUsecase
	tu *thread.ThreadUsecase
}

func NewVoteHandler(vu *usecase.VoteUsecase, tu *thread.ThreadUsecase) *VoteHandler {
	return &VoteHandler{
		vu: vu,
		tu: tu,
	}
}

func (vh *VoteHandler) Routing(r *mux.Router) {
	r.HandleFunc("/thread/{slug_or_id}/vote", http.HandlerFunc(vh.CreateVote)).Methods(http.MethodPost)
	//r.HandleFunc(`/forum/{slug}/threads`, http.HandlerFunc(th.GetAllThreadsInForum)).Methods(http.MethodGet)
}

func (vh *VoteHandler) CreateVote(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	vote := &models.Vote{}

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(err)
		return
	}

	slug_or_id := mux.Vars(r)["slug_or_id"]

	thread, createErr := vh.vu.CreateBySlugOrId(vote, slug_or_id)
	if createErr == myerror.NotExist {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(createErr)
		return
	}
	if createErr == myerror.ConflictError {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(createErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(thread)
}

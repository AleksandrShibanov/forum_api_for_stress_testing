package delivery

import (
	"encoding/json"
	"net/http"

	"forum/internal/pkg/service/usecase"

	"github.com/gorilla/mux"
)

type ServiceHandler struct {
	su *usecase.ServiceUsecase
}

func NewServiceHandler(su *usecase.ServiceUsecase) *ServiceHandler {
	return &ServiceHandler{
		su: su,
	}
}

func (sh *ServiceHandler) Routing(r *mux.Router) {
	s := r.PathPrefix("/service").Subrouter()
	s.HandleFunc("/status", http.HandlerFunc(sh.GetStatus)).Methods(http.MethodGet)
	s.HandleFunc("/clear", http.HandlerFunc(sh.ClearAll)).Methods(http.MethodPost)
	//s.HandleFunc("/{nickname}/profile", http.HandlerFunc(uh.Profile)).Methods(http.MethodGet)
	//s.HandleFunc("/{nickname}/profile", http.HandlerFunc(uh.Update)).Methods(http.MethodPost)
}

func (sh *ServiceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	status, err := sh.su.GetStatus()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (sh *ServiceHandler) ClearAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	err := sh.su.Clear()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}

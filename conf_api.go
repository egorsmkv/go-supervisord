package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type ConfAPI struct {
	router     *mux.Router
	supervisor *Supervisor
}

// NewConfAPI creates a ConfAPI object
func NewConfAPI(supervisor *Supervisor) *ConfAPI {
	return &ConfAPI{router: mux.NewRouter(), supervisor: supervisor}
}

// CreateHandler creates http handlers to process the program stdout and stderr through http interface
func (ca *ConfAPI) CreateHandler() http.Handler {
	ca.router.HandleFunc("/conf/{program}", ca.getProgramConfFile).Methods("GET")
	return ca.router
}

func (ca *ConfAPI) getProgramConfFile(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if vars == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	programName := vars["program"]
	programConfigPath := getProgramConfigPath(programName, ca.supervisor)
	if programConfigPath == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := readFile(programConfigPath)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(b)
}

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	uploadRoute  = "/upload"
	requestRoute = "/request"
)

type API struct {
	server *Server
}

func NewAPI(server *Server) API {
	return API{
		server: server,
	}
}

func (api API) Routes() http.Handler {
	r := chi.NewRouter()
	// r.Use(httplog.RequestLogger(httplog.NewLogger("merkleStoreServer", httplog.Options{JSON: true})))
	r.Post(uploadRoute, api.upload)
	r.Post(requestRoute, api.request)
	return r
}

type UploadRequest struct {
	Root    string `json:"root"`
	Index   int    `json:"index"`
	Total   int    `json:"total"`
	Content []byte `json:"content"`
}

func (api API) upload(w http.ResponseWriter, r *http.Request) {
	var upload UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&upload); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	if err := api.server.Upload(upload.Root, upload.Index, upload.Total, bytes.NewReader(upload.Content)); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type RequestRequest struct {
	Root  string `json:"root"`
	Index int    `json:"index"`
}

type RequestResponse struct {
	Content []byte   `json:"content"`
	Proof   [][]byte `json:"proof"`
}

func (api API) request(w http.ResponseWriter, r *http.Request) {
	var request RequestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	reader, proof, err := api.server.Request(request.Root, request.Index)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	response := RequestResponse{
		Content: content,
		Proof:   proof.Hashes(),
	}
	RespondWithJSON(w, http.StatusOK, response)
}

func RespondWithError(w http.ResponseWriter, code int, msg interface{}) {
	var message string
	switch m := msg.(type) {
	case error:
		message = m.Error()
	case string:
		message = m
	}
	RespondWithJSON(w, code, JSONError{Error: message})
}

type JSONError struct {
	Error string `json:"error,omitempty"`
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

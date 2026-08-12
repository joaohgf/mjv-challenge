package adapter

import (
	"encoding/json"
	"net/http"

	"github.com/example/mjv-challenge/api/dto"
	"github.com/example/mjv-challenge/api/mapper"
	"github.com/example/mjv-challenge/core/usecase"
)

type HTTPServer struct {
	publishJob usecase.PublishJob
}

func NewHTTPServer(publishJob usecase.PublishJob) *HTTPServer {
	return &HTTPServer{publishJob: publishJob}
}

func (server *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("POST /jobs", server.createJob)
	return mux
}

func (server *HTTPServer) health(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusNoContent)
}

func (server *HTTPServer) createJob(writer http.ResponseWriter, request *http.Request) {
	var body dto.CreateJobRequest
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil || body.Type == "" {
		http.Error(writer, "type is required", http.StatusBadRequest)
		return
	}

	job := mapper.ToJob(body)
	if err := server.publishJob.Execute(request.Context(), job); err != nil {
		http.Error(writer, "unable to publish job", http.StatusServiceUnavailable)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(writer).Encode(map[string]string{"id": job.ID})
}

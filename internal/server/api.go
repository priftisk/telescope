package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) RoutesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	routes := s.routeTable.List()
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(routes)
}

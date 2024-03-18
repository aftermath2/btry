package handler

import (
	"net/http"

	"github.com/aftermath2/BTRY/db"
)

// WinnersResponse is the response schema of the /winners endpoint.
type WinnersResponse struct {
	Winners []db.Winner `json:"winners,omitempty"`
}

// GetWinners responds with the list of winners.
func (h *Handler) GetWinners(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	height, err := parseIntParam(query, "height", true)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	winners, err := h.db.Winners.List(uint32(height))
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := WinnersResponse{
		Winners: winners,
	}
	sendResponse(w, http.StatusOK, resp)
}

package handler

import (
	"net/http"

	"github.com/aftermath2/BTRY/db"
)

// WinnersResponse is the response schema of the /winners and /winners/history endpoints.
type WinnersResponse struct {
	Winners []db.Winner `json:"winners,omitempty"`
}

// GetWinners responds with the list of winners.
func (h *Handler) GetWinners(w http.ResponseWriter, _ *http.Request) {
	winners, err := h.db.Winners.List()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := WinnersResponse{
		Winners: winners,
	}
	sendResponse(w, http.StatusOK, resp)
}

// GetWinnersHistory responds with the list of winners from the previous lottery.
func (h *Handler) GetWinnersHistory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	from, err := parseIntParam(query, "from")
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	to, err := parseIntParam(query, "to")
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	winners, err := h.db.Winners.ListHistory(from, to)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := WinnersResponse{
		Winners: winners,
	}
	sendResponse(w, http.StatusOK, resp)
}

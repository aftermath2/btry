package handler

import (
	"net/http"

	"github.com/aftermath2/BTRY/db"
)

// PrizesResponse contains a number representing a user's total prizes.
type PrizesResponse struct {
	Prizes uint64 `json:"prizes"`
}

// WinnersResponse is the response schema of the /winners endpoint.
type WinnersResponse struct {
	Winners []db.Winner `json:"winners,omitempty"`
}

// GetPrizes returns a public key's prizes.
func (h *Handler) GetPrizes(w http.ResponseWriter, r *http.Request) {
	publicKey, err := getAuthPublicKey(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	prizes, err := h.db.Winners.GetPrizes(publicKey)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := PrizesResponse{
		Prizes: prizes,
	}
	sendResponse(w, http.StatusOK, resp)
}

// GetWinners responds with the list of winners.
func (h *Handler) GetWinners(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	height, err := parseIntParam(query, "height")
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

package handler

import (
	"net/http"
	"strconv"

	"github.com/aftermath2/BTRY/db"

	"github.com/pkg/errors"
)

// BetsResponse is the response schema of the /bets endpoint.
type BetsResponse struct {
	Bets []db.Bet `json:"bets,omitempty"`
}

// GetBets responds with the list of bets.
func (h *Handler) GetBets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	height, err := parseIntParam(query, "height", true)
	if err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	offset, err := parseIntParam(query, "offset", false)
	if err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	limit, err := parseIntParam(query, "limit", false)
	if err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	reverse := false
	reverseStr := query.Get("reverse")
	if reverseStr != "" {
		v, err := strconv.ParseBool(reverseStr)
		if err != nil {
			sendLNURLError(w, http.StatusBadRequest, errors.Wrap(err, "invalid reverse parameter"))
			return
		}
		reverse = v
	}

	bets, err := h.db.Bets.List(uint32(height), offset, limit, reverse)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	respBody := BetsResponse{
		Bets: bets,
	}
	sendResponse(w, http.StatusOK, respBody)
}

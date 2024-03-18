package handler

import (
	"net/http"
	"strconv"

	"github.com/aftermath2/BTRY/lottery"
	"github.com/pkg/errors"
)

// HeightsResponse is the response schema of the /heights endpoint.
type HeightsResponse struct {
	Heights []uint32 `json:"heights"`
}

// GetLottery endpoint handler.
func (h *Handler) GetLottery(w http.ResponseWriter, r *http.Request) {
	lotteryInfo, err := lottery.GetInfo(r.Context(), h.lnd, h.db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResponse(w, http.StatusOK, lotteryInfo)
}

// GetHeights endpoint handler.
func (h *Handler) GetHeights(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

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

	heights, err := h.db.Lotteries.ListHeights(offset, limit, reverse)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	respBody := HeightsResponse{
		Heights: heights,
	}

	sendResponse(w, http.StatusOK, respBody)
}

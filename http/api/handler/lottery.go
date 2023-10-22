package handler

import (
	"net/http"

	"github.com/aftermath2/BTRY/lottery"
)

// GetLottery endpoint handler.
func (h *Handler) GetLottery(w http.ResponseWriter, r *http.Request) {
	lotteryInfo, err := lottery.GetInfo(r.Context(), h.lnd, h.db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResponse(w, http.StatusOK, lotteryInfo)
}

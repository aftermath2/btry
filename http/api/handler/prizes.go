package handler

import "net/http"

// PrizesResponse contains a number representing a user's total prizes.
type PrizesResponse struct {
	Prizes uint64 `json:"prizes"`
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

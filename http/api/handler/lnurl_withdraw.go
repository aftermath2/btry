package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aftermath2/BTRY/crypto"
	"github.com/fiatjaf/go-lnurl"
	"github.com/pkg/errors"
)

// LNURLWithdrawFeePPM is the default fee used for withdrawals done throguh the LNURL protocol.
//
// We use a default value to make the experience smoothly and uninterrupted.
const LNURLWithdrawFeePPM = 1500

// LNURLWithdraw endpoint handler.
func (h *Handler) LNURLWithdraw(w http.ResponseWriter, r *http.Request) {
	publicKey, err := getAuthPublicKey(r)
	if err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	query := r.URL.Query()
	signature := query.Get("signature")
	if signature == "" {
		sendLNURLError(w, http.StatusBadRequest, errors.New("signature parameter missing"))
		return
	}

	if err := crypto.VerifySignature(publicKey, signature); err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	totalPrizes, err := h.db.Winners.GetPrizes(publicKey)
	if err != nil {
		sendLNURLError(w, http.StatusInternalServerError, err)
		return
	}

	fee := totalPrizes * LNURLWithdrawFeePPM / 1_000_000

	var minWithdrawableMsat, maxWithdrawableMsat int64
	if totalPrizes > 0 {
		minWithdrawableMsat = 1000
		maxWithdrawableMsat = int64(totalPrizes-fee) * 1000
	}

	callback := fmt.Sprintf("%s/withdraw?fee=%d&pubkey=%s",
		strings.TrimSuffix(r.RequestURI, r.URL.Path),
		fee,
		publicKey,
	)
	resp := &lnurl.LNURLWithdrawResponse{
		Tag:                "withdrawalRequest",
		Callback:           callback,
		K1:                 signature,
		DefaultDescription: "BTRY withdrawal",
		MinWithdrawable:    minWithdrawableMsat,
		MaxWithdrawable:    maxWithdrawableMsat,
	}
	sendResponse(w, http.StatusOK, resp)
}

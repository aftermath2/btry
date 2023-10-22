package handler

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/aftermath2/BTRY/lottery"

	"github.com/pkg/errors"
)

// InvoiceResponse is the response schema of the /invoices endpoint.
type InvoiceResponse struct {
	Invoice   string `json:"invoice,omitempty"`
	PaymentID uint64 `json:"payment_id,omitempty"`
}

// GetInvoice reponds with an invoice and its preimage hash.
func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	publicKey, err := getAuthPublicKey(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	query := r.URL.Query()

	amount := query.Get("amount")
	amountSat, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, errors.Wrap(err, "invalid amount"))
		return
	}

	ctx := r.Context()

	lotteryInfo, err := lottery.GetInfo(ctx, h.lnd, h.db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	// An invoice may be requested before the capacity has been fulfilled but pay afterwards,
	// the user would participate in the lottery but the funds may not be considered in the pool
	// (assuming the liquidity remains the same and no withdrawal is done in the same day)
	if amountSat > uint64(lotteryInfo.Capacity) {
		err := errors.Errorf(
			"requested amount exceeds current capacity. Amount should be equal or lower than %d",
			lotteryInfo.Capacity)
		sendError(w, http.StatusBadRequest, err)
		return
	}

	inv, err := h.lnd.AddInvoice(ctx, amountSat)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	rHash := hex.EncodeToString(inv.RHash)
	paymentID := h.eventStreamer.TrackPayment(rHash, publicKey, amountSat)

	resp := InvoiceResponse{
		PaymentID: paymentID,
		Invoice:   inv.PaymentRequest,
	}
	sendResponse(w, http.StatusOK, resp)
}

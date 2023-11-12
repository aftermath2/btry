package handler

import (
	"net/http"
	"time"

	"github.com/aftermath2/BTRY/crypto"

	"github.com/pkg/errors"
)

// WithdrawResponse is the response schema of the /withdraw endpoint.
type WithdrawResponse struct {
	Status    string `json:"status,omitempty"`
	PaymentID uint64 `json:"payment_id,omitempty"`
}

// Withdraw handles a withdrawal request by attempting to pay an invoice.
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	signature := query.Get("k1")
	if signature == "" {
		sendLNURLError(w, http.StatusBadRequest, errors.New("k1 parameter missing"))
		return
	}

	publicKey := query.Get("pubkey")
	if publicKey == "" {
		sendLNURLError(w, http.StatusBadRequest, errors.New("pubkey parameter missing"))
		return
	}

	if err := crypto.VerifySignature(publicKey, signature); err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	paymentRequest := query.Get("pr")
	if paymentRequest == "" {
		sendLNURLError(w, http.StatusBadRequest, errors.New("pr parameter missing"))
		return
	}

	fee, err := parseIntParam(query, "fee")
	if err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()

	invoice, err := h.lnd.DecodeInvoice(ctx, paymentRequest)
	if err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	if invoice.NumSatoshis == 0 {
		sendLNURLError(w, http.StatusBadRequest, errors.New("invalid invoice amount"))
		return
	}

	if time.Now().Unix() >= (invoice.Timestamp + invoice.Expiry) {
		sendLNURLError(w, http.StatusBadRequest, errors.New("invoice expired"))
		return
	}

	withdrawAmount := uint64(invoice.NumSatoshis) + fee

	// Here the invoice amount is deducted from the public key prize and persisted, if the payment
	// fails, the user will get its funds restored.
	// It's done this way to not let users request more funds than they have.
	if err := h.db.Winners.ClaimPrizes(publicKey, withdrawAmount); err != nil {
		sendLNURLError(w, http.StatusBadRequest, err)
		return
	}

	paymentID := h.eventStreamer.TrackPayment(invoice.PaymentHash, publicKey, withdrawAmount)

	if _, err := h.lnd.PayInvoice(ctx, invoice, int64(fee), false); err != nil {
		sendLNURLError(w, http.StatusInternalServerError, err)
		return
	}

	resp := WithdrawResponse{
		PaymentID: paymentID,
		Status:    "OK",
	}
	sendResponse(w, http.StatusOK, resp)
}

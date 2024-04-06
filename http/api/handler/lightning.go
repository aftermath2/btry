package handler

import (
	"net/http"

	"github.com/aftermath2/BTRY/db"

	"github.com/fiatjaf/go-lnurl"
	"github.com/pkg/errors"
)

// GetLightningAddressResponse is the response schema of the GET /lightning/address endpoint.
type GetLightningAddressResponse struct {
	Address    string `json:"address,omitempty"`
	HasAddress bool   `json:"has_address,omitempty"`
}

// SetLightningAddressResponse is the response schema of the POST /lightning/address endpoint.
type SetLightningAddressResponse struct {
	Success bool `json:"success,omitempty"`
}

// GetLightningAddress responds with the public key's linked lightning address.
func (h *Handler) GetLightningAddress(w http.ResponseWriter, r *http.Request) {
	publicKey, err := getAuthPublicKey(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	address, err := h.db.Lightning.GetAddress(publicKey)
	if err != nil {
		if errors.Is(err, db.ErrNoAddress) {
			resp := GetLightningAddressResponse{HasAddress: false}
			sendResponse(w, http.StatusOK, resp)
			return
		}
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := GetLightningAddressResponse{
		Address:    address,
		HasAddress: true,
	}
	sendResponse(w, http.StatusOK, resp)
}

// SetLightningAddress links a public key with a lightning address.
func (h *Handler) SetLightningAddress(w http.ResponseWriter, r *http.Request) {
	publicKey, err := getAuthPublicKey(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	query := r.URL.Query()
	address := query.Get("address")
	if _, _, ok := lnurl.ParseInternetIdentifier(address); !ok {
		sendError(w, http.StatusBadRequest, errors.New("invalid lightning address"))
		return
	}

	if err := h.db.Lightning.SetAddress(publicKey, address); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := SetLightningAddressResponse{Success: true}
	sendResponse(w, http.StatusOK, resp)
}

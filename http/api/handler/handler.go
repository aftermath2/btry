package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aftermath2/BTRY/crypto"
	"github.com/aftermath2/BTRY/db"
	"github.com/aftermath2/BTRY/http/api/sse"
	"github.com/aftermath2/BTRY/lightning"

	"github.com/fiatjaf/go-lnurl"
	"github.com/pkg/errors"
)

// ErrorResponse is the object returned when an error is thrown.
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// Handler handles endpoints requests.
type Handler struct {
	lnd           lightning.Client
	db            *db.DB
	eventStreamer sse.Streamer
}

// New returns the endpoints handler.
func New(lnd lightning.Client, db *db.DB, eventStreamer sse.Streamer) *Handler {
	return &Handler{
		lnd:           lnd,
		db:            db,
		eventStreamer: eventStreamer,
	}
}

func getAuthPublicKey(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	splitPubKey := strings.Split(auth, "Bearer ")
	if len(splitPubKey) != 2 {
		return "", errors.New("invalid authorization public key")
	}
	publicKey := splitPubKey[1]

	if err := crypto.ValidatePublicKey(publicKey); err != nil {
		return "", err
	}

	return publicKey, nil
}

func parseIntParam(query url.Values, key string, required bool) (uint64, error) {
	str := query.Get(key)
	if str == "" {
		if required {
			return 0, errors.Errorf("query parameter %q is required", key)
		}
		return 0, nil
	}

	n, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "invalid %s", key)
	}

	return n, nil
}

func sendResponse(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed encoding response body", http.StatusInternalServerError)
	}
}

func sendLNURLError(w http.ResponseWriter, statusCode int, err error) {
	sendResponse(w, statusCode, lnurl.ErrorResponse(err.Error()))
}

func sendError(w http.ResponseWriter, statusCode int, err error) {
	errResponse := ErrorResponse{
		Error: err.Error(),
	}
	sendResponse(w, statusCode, errResponse)
}

package lightning

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/pkg/errors"
)

const (
	// HODLInvoiceExpiry represents HODL invoices expiry in seconds
	//
	// Battles have the same expiration as a consequence. 3 hours.
	HODLInvoiceExpiry int64 = 10_800

	hodlContractBaseFee = 20
	// Fee to take part in a battle is holdContractFeePPM - settlementFeePPM
	hodlContractFeePPM = 1000
	settlementFeePPM   = hodlContractFeePPM - 300

	cltvDelta     = 48
	minCLTVExpiry = 200
	// Should be set to at most the node's `--max-cltv-expiry` setting (default: 2016)
	maxCLTVExpiry = 2016
)

// AddHODLInvoice takes an invoice from a user and creates a HODL invoice with that invoice's
// payment hash, we do not have custody of the funds since the preimage is secret and only known
// by the invoice creator.
func (c *client) AddHODLInvoice(ctx context.Context, invoice *lnrpc.PayReq) (string, error) {
	if invoice.Expiry < HODLInvoiceExpiry {
		return "", errors.Errorf("invoice expiry is too short, should be more than %v", time.Duration(HODLInvoiceExpiry))
	}

	paymentHash, err := hex.DecodeString(invoice.PaymentHash)
	if err != nil {
		return "", errors.Wrap(err, "decoding payment hash")
	}

	fee := hodlContractBaseFee + (invoice.NumSatoshis * hodlContractFeePPM / 1_000_000)
	cltvExpiry := invoice.CltvExpiry + (cltvDelta * 2)
	if cltvExpiry > maxCLTVExpiry {
		return "", errors.New("CLTV expiry too long")
	}
	if cltvExpiry < minCLTVExpiry {
		cltvExpiry = minCLTVExpiry
	}

	hodlInvoice, err := c.invoices.AddHoldInvoice(ctx, &invoicesrpc.AddHoldInvoiceRequest{
		Memo:       "BTRY HODL contract",
		Hash:       paymentHash,
		Value:      invoice.NumSatoshis + fee,
		Expiry:     HODLInvoiceExpiry,
		CltvExpiry: uint64(cltvExpiry),
	})
	if err != nil {
		return "", err
	}

	return hodlInvoice.PaymentRequest, nil
}

// CancelInvoice cancels a currently open invoice.
// If the invoice is already cancelled, this call will succeed.
// If the invoice is already settled, it will fail.
func (c *client) CancelInvoice(ctx context.Context, rHash string) error {
	paymentHash, err := hex.DecodeString(rHash)
	if err != nil {
		return errors.Wrap(err, "decoding payment hash")
	}

	_, err = c.invoices.CancelInvoice(ctx, &invoicesrpc.CancelInvoiceMsg{
		PaymentHash: paymentHash,
	})
	if err != nil {
		return errors.Wrap(err, "canceling invoice")
	}

	return nil
}

// SettleInvoice settles an accepted invoice.
// If the invoice is already settled, this call will succeed.
func (c *client) SettleInvoice(ctx context.Context, originalInvoice string) error {
	payReq, err := c.DecodeInvoice(ctx, originalInvoice)
	if err != nil {
		return err
	}

	paymentHash, err := hex.DecodeString(payReq.PaymentHash)
	if err != nil {
		return errors.Wrap(err, "decoding payment hash")
	}

	invoice, err := c.ln.LookupInvoice(ctx, &lnrpc.PaymentHash{RHash: paymentHash})
	if err != nil {
		return err
	}

	// The invoice should already been accepted by the other party,
	// otherwise we risk making the payment
	if invoice.State != lnrpc.Invoice_ACCEPTED {
		if invoice.State != lnrpc.Invoice_CANCELED {
			if err := c.CancelInvoice(ctx, payReq.PaymentHash); err != nil {
				return err
			}
		}
		return errors.Errorf("invoice state is not %s", lnrpc.Invoice_ACCEPTED)
	}

	// Payment CLTV <-- cltvDelta * 3 --> HOLD invoice CLTV
	cltvLimit := int32(invoice.CltvExpiry - cltvDelta)
	fee := payReq.NumSatoshis * settlementFeePPM / 1_000_000

	stream, err := c.PayInvoice(ctx, payReq, fee, cltvLimit)
	if err != nil {
		return err
	}

	payment, err := stream.Recv()
	if err != nil {
		return err
	}

	switch payment.Status {
	case lnrpc.Payment_FAILED:
		if err := c.CancelInvoice(ctx, payReq.PaymentHash); err != nil {
			return err
		}

		return errors.Errorf("paying %s invoice", payment.PaymentHash)

	case lnrpc.Payment_SUCCEEDED:
		preimage, err := hex.DecodeString(payment.PaymentPreimage)
		if err != nil {
			return errors.Wrap(err, "decoding preimage")
		}

		_, err = c.invoices.SettleInvoice(ctx, &invoicesrpc.SettleInvoiceMsg{
			Preimage: preimage,
		})
		if err != nil {
			return errors.Wrap(err, "settling invoice")
		}
	}

	return nil
}

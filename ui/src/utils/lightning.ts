import { bech32 } from "@scure/base";
// @ts-ignore
import { decode } from "light-bolt11-decoder";

const errInvalidNetwork = Error(`network must be ${import.meta.env.VITE_NETWORK}`)
const errExpiredInvoice = Error("already expired")
const errInvalidAmount = Error("invalid amount")
const errAmountTooHigh = Error("amount is higher than available prizes")

export interface Invoice {
	paymentHash: string
	amountSat: number
}

/**
 * LNURLEncode takes a normal URL and returns its LNURL encoded version
 * 
 * @param url plain URL
 * @returns LNURL
 */
export const LNURLEncode = (url: string): string => {
	const words = bech32.toWords(new TextEncoder().encode(url))
	return bech32.encode("lnurl", words, 1023)
}

/**
 * ValidateInvoice throws an error if the provided invoice is invalid.
 * 
 * @param payReq payment request
 * @param targetAmount exact invoice amount in sats expected
 * @param maxAmount maximum invoice amount in sats expected
 */
export const ValidateInvoice = (payReq: string, targetAmount?: number, maxAmount?: number): Invoice => {
	if (!payReq.startsWith(import.meta.env.VITE_BOLT11_PREFIX)) {
		throw errInvalidNetwork
	}

	try {
		const decodedInvoice = decode(payReq);
		const invoice: Invoice = {
			amountSat: decodedInvoice.sections[2].value / 1000,
			paymentHash: decodedInvoice.payment_hash
		}

		if (!invoice.amountSat) {
			throw errInvalidAmount
		}

		if (targetAmount && invoice.amountSat !== targetAmount) {
			throw errInvalidAmount
		}

		if (maxAmount && invoice.amountSat > maxAmount) {
			throw errAmountTooHigh
		}

		const expireDate = decodedInvoice.sections[4].value + decodedInvoice.expiry
		if (expireDate <= Math.floor(Date.now() / 1000)) {
			throw errExpiredInvoice
		}

		return invoice
	} catch (error) {
		throw error
	}
}

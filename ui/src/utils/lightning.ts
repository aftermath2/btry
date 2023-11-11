import { bech32 } from "@scure/base";
// @ts-ignore
import { decode } from "light-bolt11-decoder";

const errInvalidNetwork = Error(`invalid invoice network, it must be ${import.meta.env.VITE_NETWORK}`)
const errExpiredInvoice = Error("invoice already expired")
const errInvalidAmount = Error("invalid invoice amount")
const errAmountTooHigh = Error("invoice amount is higher than available prizes")

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
 * @param invoice payment request
 * @param targetAmount exact invoice amount in sats expected
 * @param maxAmount maximum invoice amount in sats expected
 */
export const ValidateInvoice = (invoice: string, targetAmount?: number, maxAmount?: number) => {
	if (!invoice.startsWith(import.meta.env.VITE_BOLT11_PREFIX)) {
		throw errInvalidNetwork
	}

	try {
		const decodedInvoice = decode(invoice);

		const amountSat = Math.round(decodedInvoice.sections[2].value / 1000);
		if (!amountSat) {
			throw errInvalidAmount
		}

		if (targetAmount && amountSat !== targetAmount) {
			throw errInvalidAmount
		}

		if (maxAmount && amountSat > maxAmount) {
			throw errAmountTooHigh
		}

		const expireDate = decodedInvoice.sections[4].value + decodedInvoice.expiry
		if (expireDate <= Math.floor(Date.now() / 1000)) {
			throw errExpiredInvoice
		}
	} catch (error) {
		throw error
	}
}

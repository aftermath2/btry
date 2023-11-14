import { etc } from "@noble/ed25519";

export const NumberRegex = /[0-9]/
export const HexRegex = /^[0-9a-fA-F]+$/

/**
 * BeatufiyNumber formats a number in a way that it's easier to read.
 * 
 * @param n number
 * @returns string with the number having spaces every three digits
 */
export const BeautifyNumber = (n?: number, sep?: string): string => {
	if (!n) {
		return "0"
	}
	const numStr = n.toLocaleString("en-US")
	if (!sep) {
		return numStr
	}
	return numStr.replaceAll(",", sep)
}

/**
 * FormatTime converts HH:MM:SS into the desired format.
 * 
 * @param t time value
 * @returns formatted time
 */
export const FormatTime = (t: string): string => {
	// Already matches the format
	if (t.length === 8) {
		return t
	}

	const parts = t.split(":")
	parts.forEach((part, index, self) => {
		if (part.length === 1) {
			self[index] = "0" + part
		}
	})
	return parts.join(":")
}


/**
 * Hash takes any string and returns its SHA-256 hash.
 * 
 * @param key string to hash
 * @returns hex-encoded SHA-256 hash of the key
 */
export const Hash = async (key: string): Promise<string> => {
	const buf = new TextEncoder().encode(key)
	const hash = await crypto.subtle.digest("SHA-256", buf)
	const hashArray = new Uint8Array(hash)
	return etc.bytesToHex(hashArray)
}

/**
 * HexEncode encodes text using the hexadecimal numeral system. 
 * 
 * @param text raw string
 * @returns hex-encoded string
 */
export const HexEncode = (text: string): string => {
	return text.split("")
		.map(c => c.charCodeAt(0).toString(16).padStart(2, "0"))
		.join("");
}

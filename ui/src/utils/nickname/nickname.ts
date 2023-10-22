import { BigNumber } from 'bignumber.js';

import { adjectives } from "./adjectives";
import { nouns } from "./nouns";
import { Hash } from "../utils";

const maxNum = 999
const maxHashNum = BigNumber(2 ** 256)
/** 1000 * 4800 * 12500 = 60 Billion deterministic nicks */
const poolSize = maxNum * adjectives.length * nouns.length

/**
 * NicknameFromKey returns a deterministic nickname from a hex-encoded SHA-256 hash or a key
 * of the same length.
 * 
 * Credits to Robosats for the implementation.
 * 
 * @param key SHA-256 hash or 32 character long key
 * @returns A deterministic nickname
 */
export const NicknameFromKey = async (key: string): Promise<string> => {
	// Key must be 64 characters long (32 hex-encoded bytes)
	if (key.length !== 64) {
		key = await Hash(key)
	}

	const keyNum = BigNumber(parseInt(key, 16))
	const nickID = Math.floor(keyNum.div(maxHashNum, 10).toNumber() * poolSize)
	let remainder = nickID

	// Adjective
	const adjectiveID = Math.floor(remainder / (maxNum * nouns.length))
	const adjective = adjectives[adjectiveID]
	remainder = remainder - adjectiveID * maxNum * nouns.length

	// Noun
	const nounID = Math.floor(remainder / maxNum)
	const noun = nouns[nounID]

	// Number
	const num = remainder - nounID * maxNum

	return adjective + noun + num.toString()
}
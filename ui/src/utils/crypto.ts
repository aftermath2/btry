import * as ed from '@noble/ed25519';

/**
 * GenerateKeyPair returns ed25519 private-public key pair.
 * 
 * @returns [privateKey, publicKey]
 */
export const GenerateKeyPair = async (): Promise<[string, string]> => {
	const privateKey = ed.utils.randomPrivateKey()
	const publicKey = await ed.getPublicKeyAsync(privateKey)
	return [ed.etc.bytesToHex(privateKey), ed.etc.bytesToHex(publicKey)]
}

/**
 * GetPubicKey returns a private key's corresponding public key.
 * 
 * @param privateKey hex-encoded private key, must be 64 charaters long
 * @returns hex-encoded public key
 */
export const GetPublicKey = async (privateKey: string): Promise<string> => {
	const privKey = ed.etc.hexToBytes(privateKey)
	const publicKey = await ed.getPublicKeyAsync(privKey)
	return ed.etc.bytesToHex(publicKey)
}

/**
 * Sign takes a public key and signs it using the private key.
 * 
 * @param publicKey user public key
 * @param privateKey user private key
 * @returns hex-encoded signature of the public key
 */
export const Sign = async (privateKey: string, publicKey: string): Promise<string> => {
	const signature = await ed.signAsync(publicKey, privateKey)
	return ed.etc.bytesToHex(signature)
}
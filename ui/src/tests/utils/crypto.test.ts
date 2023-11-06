import { GenerateKeyPair, GetPublicKey, Sign } from "../../utils/crypto";
import { HexRegex } from "../../utils/utils";

describe('GenerateKeyPair', () => {
	test('Validate ED25519 key pair', async () => {
		const [privateKey, publicKey] = await GenerateKeyPair()

		// Validate keys length
		const expectedLength = 64
		expect(publicKey.length).toBe(expectedLength)
		expect(privateKey.length).toBe(expectedLength)

		// Validate hex encoding
		expect(HexRegex.test(publicKey)).toBeTruthy()
		expect(HexRegex.test(publicKey)).toBeTruthy()
	});
});

describe('GetPublicKey', () => {
	test('Obtain public key from private key', async () => {
		const privateKey = "567dab3c2385a3bd0ce03ca67466cdebedeabfc553588cc31a09bcd77efd4360"
		const publicKey = "3eda780a103cb7038cfad4de468bdc9532de1c223a6e8fb1105adef362c306f6"

		expect(await GetPublicKey(privateKey)).toBe(publicKey);
	});
});

describe('Sign', () => {
	test('Validate signature', async () => {
		const privateKey = "0a20cec75e014c4afb5bccbd194b20e6fea7c727a3ccfdf6b72227154a575343"
		const publicKey = "6281adefcbf753053863061d414a905fb5b9063c22ec44feea10d82c8793a9a9"
		const signature = "52ac82a03b6fc0d7e4f2aedcaf2b792dbcbd11b1e2664e4659f24429d0b01dbf5707ecb678121c31c2fe6586318c674fe8d04cbf14e194fe461a8dfcb2f7d903"

		expect(await Sign(privateKey, publicKey)).toBe(signature);
	});
});
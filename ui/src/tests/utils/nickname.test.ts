import { NicknameFromKey } from "../../utils/nickname/nickname";

describe('NicknameFromKey', () => {
	test('ED25519 public key', async () => {
		const publicKey = "f73a68a323bca465ad2dd2cab8df07d5bb854984277d6354e3c16764e9ddd079"
		expect(await NicknameFromKey(publicKey)).toBe("OneInscrutability632");
	});

	test('Custom string', async () => {
		expect(await NicknameFromKey("test")).toBe("SuccessfulSophomore682");
	});
});
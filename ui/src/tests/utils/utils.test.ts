import { BeautifyNumber, Hash, HexEncode, HexRegex } from "../../utils/utils";

describe("BeautifyNumber", () => {
	test("Zero", () => {
		expect(BeautifyNumber(0)).toBe("0")
	})

	test("Default separator", () => {
		expect(BeautifyNumber(100_000)).toBe("100,000")
	})

	test("Custom separator", () => {
		expect(BeautifyNumber(5_000_000, "_")).toBe("5_000_000")
	})
})

describe("Hash", () => {
	test("Verify that some text is hashed with SHA-256 as expected", async () => {
		const expected = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
		expect(await Hash("test")).toBe(expected)
	});
});

describe("HexEncode", () => {
	test("Verify that text is hex encoded correctly", () => {
		const result = HexEncode("test")
		expect(result).toBe("74657374")

		expect(HexRegex.test(result)).toBeTruthy()
	});
});

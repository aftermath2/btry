import { dict } from "../../i18n/i18n";

describe("i18n dictionary", () => {
	test("All languages have the same keys", () => {
		const languages = Object.entries(dict);

		for (const [lang, entries] of languages) {
			for (const [alt, altEntries] of languages) {
				if (lang === alt) {
					continue;
				}

				const langKeys = new Set(Object.keys(entries));
				const altKeys = new Set(Object.keys(altEntries));

				altKeys.forEach((value) => langKeys.delete(value));

				expect(Array.from(langKeys.values())).toEqual([]);
			}
		}
	});
});
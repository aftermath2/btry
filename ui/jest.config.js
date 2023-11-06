module.exports = {
	transform: {
		"^.+\\.(t|j)s$": ["ts-jest", { useESM: true, }]
	},
	extensionsToTreatAsEsm: [".ts"],
	moduleNameMapper: {
		"^(\\.{1,2}/.*)\\.js$": "$1",
	},
	testEnvironment: "node",
	testRegex: "tests/.*\\.(test|spec)?\\.(ts|tsx)$",
	moduleFileExtensions: ["js", "json", "ts"],
	rootDir: "src",
	extensionsToTreatAsEsm: [".ts"],
	transformIgnorePatterns: ["/node_modules/(?!@noble)"],
};
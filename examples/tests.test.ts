// Unit Testing Example
// Run with: gots test
// Note: Test functions would be discovered and executed by the test runner

// Test suite: Calculator
const calculatorTests = {
    "should add two numbers": () => {
        const sum = 2 + 3;
        console.assert(sum === 5, "Addition failed");
    },
    "should handle negative numbers": () => {
        const sum = -2 + 3;
        console.assert(sum === 1, "Negative addition failed");
    },
    "should handle zero": () => {
        const sum = 0 + 5;
        console.assert(sum === 5, "Zero addition failed");
    },
};

// Test suite: Subtraction
const subtractionTests = {
    "should subtract two numbers": () => {
        const diff = 5 - 3;
        console.assert(diff === 2, "Subtraction failed");
    },
    "should handle negative results": () => {
        const diff = 2 - 5;
        console.assert(diff === -3, "Negative result failed");
    },
};

// Test suite: String Operations
const stringTests = {
    "should concatenate strings": () => {
        const result = "Hello" + " " + "World";
        console.assert(result === "Hello World", "Concatenation failed");
    },
    "should handle empty strings": () => {
        const result = "" + "test";
        console.assert(result === "test", "Empty string concat failed");
    },
    "should find substring": () => {
        const text = "Hello World";
        console.assert(text.includes("World"), "Substring not found");
    },
};

// Test suite: Array Operations
const arrayTests = {
    "should map array": () => {
        const arr = [1, 2, 3];
        const mapped = arr.map((x) => x * 2);
        console.assert(mapped.length === 3, "Map failed");
    },
    "should filter array": () => {
        const arr = [1, 2, 3, 4, 5];
        const filtered = arr.filter((x) => x > 2);
        console.assert(filtered.length === 3, "Filter failed");
    },
};

// Helper test runner
function runTests(): void {
    let passed = 0;
    let failed = 0;

    const suites: Record<string, Record<string, () => void>> = {
        "Calculator": calculatorTests,
        "Subtraction": subtractionTests,
        "String Operations": stringTests,
        "Array Operations": arrayTests,
    };

    for (const suiteName in suites) {
        const suite = suites[suiteName];
        console.log(`\n${suiteName}:`);

        for (const testName in suite) {
            try {
                suite[testName]();
                console.log(`  ✓ ${testName}`);
                passed++;
            } catch (error) {
                console.log(`  ✗ ${testName}: ${error}`);
                failed++;
            }
        }
    }

    console.log(`\nTests: ${passed} passed, ${failed} failed`);
}

// Run tests
runTests();

export { calculatorTests, subtractionTests, stringTests, arrayTests };


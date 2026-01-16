// Async/Await Example
// Run with: gots run async.ts

async function fetchData(url: string): Promise<string> {
    // Simulated async operation
    return new Promise((resolve) => {
        setTimeout(() => {
            resolve(`Data from ${url}`);
        }, 1000);
    });
}

async function processData(): Promise<void> {
    console.log("Starting async operations...");

    try {
        const data1 = await fetchData("http://api.example.com/users");
        console.log("Result 1:", data1);

        const data2 = await fetchData("http://api.example.com/posts");
        console.log("Result 2:", data2);

        // Parallel execution
        const [result1, result2] = await Promise.all([
            fetchData("http://api.example.com/comments"),
            fetchData("http://api.example.com/tags"),
        ]);

        console.log("Parallel Result 1:", result1);
        console.log("Parallel Result 2:", result2);

        console.log("All operations completed!");
    } catch (error) {
        console.error("Error during async operations:", error);
    }
}

// Main execution
async function main(): Promise<void> {
    await processData();
    console.log("Program completed");
}

main().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
});

export { processData, fetchData };

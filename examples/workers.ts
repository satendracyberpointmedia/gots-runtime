// Concurrency Example with Workers
// Run with: gots run workers.ts

import { createWorkerPool, spawnWorker, createTask } from "../stdlib/worker/index";

interface WorkerInput {
    id: number;
    value: number;
}

interface WorkerOutput {
    id: number;
    result: number;
}

// Worker task: expensive computation
const computeTask = createTask<WorkerInput, WorkerOutput>(
    { id: 1, value: 10 },
    async (input) => {
        // Simulate expensive computation
        let result = 0;
        for (let i = 0; i < input.value * 1000000; i++) {
            result += i % 2 === 0 ? 1 : -1;
        }
        return { id: input.id, result };
    }
);

async function runConcurrentTasks(): Promise<void> {
    console.log("Starting worker pool with 4 workers...");

    const pool = createWorkerPool(2, 4);
    const tasks: Promise<any>[] = [];

    // Submit 10 tasks
    for (let i = 0; i < 10; i++) {
        const task = createTask<WorkerInput, WorkerOutput>(
            { id: i, value: 5 + (i % 3) },
            async (input) => {
                let result = 0;
                for (let j = 0; j < input.value * 100000; j++) {
                    result += j % 2 === 0 ? 1 : -1;
                }
                return { id: input.id, result };
            }
        );

        tasks.push(spawnWorker(task));
    }

    console.log(`Submitted ${tasks.length} tasks to worker pool`);

    try {
        const results = await Promise.all(tasks);
        console.log("Completed tasks:", results.length);
        results.forEach((result) => {
            console.log(`Task ${result.id}: ${result.result}`);
        });
    } catch (error) {
        console.error("Worker error:", error);
    }

    console.log("Worker pool shutdown complete");
}

// Run
runConcurrentTasks().catch(console.error);

export { runConcurrentTasks };

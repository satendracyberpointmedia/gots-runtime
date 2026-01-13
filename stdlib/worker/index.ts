// Standard Library: Worker Threads
// TypeScript definitions for zero-cost worker thread abstraction

export interface WorkerTask<T = any, R = any> {
    id: string;
    data: T;
    handler?: (data: T) => R | Promise<R>;
}

export interface WorkerResult<T = any> {
    id: string;
    data: T;
    error?: Error;
    duration: number; // milliseconds
}

export interface Worker {
    spawn<T, R>(task: WorkerTask<T, R>): Promise<WorkerResult<R>>;
    spawnBatch<T, R>(tasks: WorkerTask<T, R>[]): Promise<WorkerResult<R>[]>;
    close(): Promise<void>;
}

export interface WorkerPool {
    spawn<T, R>(task: WorkerTask<T, R>): Promise<WorkerResult<R>>;
    spawnBatch<T, R>(tasks: WorkerTask<T, R>[]): Promise<WorkerResult<R>[]>;
    getStats(): {
        totalWorkers: number;
        busyWorkers: number;
        idleWorkers: number;
        queuedTasks: number;
    };
}

// Factory function to create a worker pool
export function createWorkerPool(minWorkers?: number, maxWorkers?: number): WorkerPool;

// Factory function to spawn a single worker
export function spawnWorker<T, R>(task: WorkerTask<T, R>): Promise<WorkerResult<R>>;


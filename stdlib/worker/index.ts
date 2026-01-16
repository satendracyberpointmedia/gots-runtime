// Standard Library: Worker Threads
// TypeScript definitions for zero-cost worker thread abstraction

export interface WorkerTask<T = any, R = any> {
    id: string;
    data: T;
    handler?: (data: T) => R | Promise<R>;
    priority?: number; // 0-10, higher = more important
    timeout?: number; // milliseconds
}

export interface WorkerResult<T = any> {
    id: string;
    data: T;
    error?: Error;
    duration: number; // milliseconds
    workerId: string;
}

export interface Worker {
    spawn<T, R>(task: WorkerTask<T, R>): Promise<WorkerResult<R>>;
    spawnBatch<T, R>(tasks: WorkerTask<T, R>[]): Promise<WorkerResult<R>[]>;
    close(): Promise<void>;
    isAvailable(): boolean;
}

export interface WorkerPool {
    spawn<T, R>(task: WorkerTask<T, R>): Promise<WorkerResult<R>>;
    spawnBatch<T, R>(tasks: WorkerTask<T, R>[]): Promise<WorkerResult<R>[]>;
    spawnWithPriority<T, R>(task: WorkerTask<T, R>, priority: number): Promise<WorkerResult<R>>;

    getStats(): {
        totalWorkers: number;
        busyWorkers: number;
        idleWorkers: number;
        queuedTasks: number;
        completedTasks: number;
        failedTasks: number;
    };

    resize(minWorkers: number, maxWorkers: number): Promise<void>;
    shutdown(): Promise<void>;
    warmUp(count: number): Promise<void>;
}

// Factory function to create a worker pool
export function createWorkerPool(minWorkers?: number, maxWorkers?: number): WorkerPool { throw new Error('Not implemented'); }

// Factory function to spawn a single worker
export function spawnWorker<T, R>(task: WorkerTask<T, R>): Promise<WorkerResult<R>> { throw new Error('Not implemented'); }

// Utility functions
export function isWorkerTask(obj: any): boolean { throw new Error('Not implemented'); }
export function createTask<T, R>(data: T, handler: (data: T) => R | Promise<R>, id?: string): WorkerTask<T, R> { throw new Error('Not implemented'); }

// Standard Library: Profiler
// TypeScript definitions for production-grade profiling tools

export interface ProfilerResult {
    cpu?: string;      // CPU profile data (base64 encoded)
    memory?: string;   // Memory profile data (base64 encoded)
    goroutine?: string; // Goroutine profile data (base64 encoded)
    block?: string;    // Block profile data (base64 encoded)
    mutex?: string;    // Mutex profile data (base64 encoded)
}

export interface Profiler {
    startCPU(outputPath?: string): Promise<void>;
    startMemory(outputPath?: string): Promise<void>;
    startGoroutine(outputPath?: string): Promise<void>;
    startBlock(outputPath?: string): Promise<void>;
    startMutex(outputPath?: string): Promise<void>;
    stop(): Promise<ProfilerResult>;
    getResults(): ProfilerResult;
}

// Factory function to get profiler
export function getProfiler(): Profiler;


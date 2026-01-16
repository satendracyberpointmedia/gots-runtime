// Standard Library: Native RPC System
// TypeScript definitions for native RPC communication

export interface RPCRequest {
    id: string;
    method: string;
    params?: any;
    module?: string;
    timeout?: number;
}

export interface RPCResponse {
    id: string;
    result?: any;
    error?: RPCError;
    duration?: number;
}

export interface RPCError {
    code: number;
    message: string;
    data?: any;
}

export type RPCHandler = (params: any, context?: RPCContext) => Promise<any> | any;

export interface RPCContext {
    clientId: string;
    method: string;
    requestId: string;
    metadata: Record<string, any>;
}

export interface RPCServer {
    register(method: string, handler: RPCHandler): RPCServer;
    unregister(method: string): RPCServer;
    registerModule(moduleName: string, handlers: Record<string, RPCHandler>): RPCServer;

    listen(address: string, callback?: (err?: Error) => void): void;
    close(callback?: (err?: Error) => void): void;

    getStats(): {
        registeredMethods: number;
        activeConnections: number;
        totalRequests: number;
        totalErrors: number;
    };
}

export interface RPCClient {
    call(method: string, params?: any, timeout?: number): Promise<any>;
    callModule(module: string, method: string, params?: any, timeout?: number): Promise<any>;
    batch(calls: Array<{ method: string, params?: any }>): Promise<any[]>;
    close(): Promise<void>;
    isConnected(): boolean;
}

export interface RPCServerOptions {
    maxConnections?: number;
    requestTimeout?: number;
    maxRequestSize?: number;
    allowedMethods?: string[];
}

export interface RPCClientOptions {
    timeout?: number;
    retries?: number;
    retryDelay?: number;
    keepAlive?: boolean;
}

// Factory functions
export function createServer(options?: RPCServerOptions): RPCServer { throw new Error('Not implemented'); }
export function createClient(address: string, options?: RPCClientOptions): Promise<RPCClient> { throw new Error('Not implemented'); }

// Utility functions
export function isRPCRequest(obj: any): boolean { throw new Error('Not implemented'); }
export function isRPCResponse(obj: any): boolean { throw new Error('Not implemented'); }
export function createRPCError(code: number, message: string, data?: any): RPCError { throw new Error('Not implemented'); }

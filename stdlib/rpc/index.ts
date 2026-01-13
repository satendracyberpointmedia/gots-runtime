// Standard Library: Native RPC System
// TypeScript definitions for native RPC communication

export interface RPCRequest {
    id: string;
    method: string;
    params?: any;
    module?: string;
}

export interface RPCResponse {
    id: string;
    result?: any;
    error?: RPCError;
}

export interface RPCError {
    code: number;
    message: string;
    data?: any;
}

export type RPCHandler = (params: any) => Promise<any> | any;

export interface RPCServer {
    register(method: string, handler: RPCHandler): void;
    unregister(method: string): void;
    listen(address: string, callback?: (err?: Error) => void): void;
    close(callback?: (err?: Error) => void): void;
}

export interface RPCClient {
    call(method: string, params?: any): Promise<any>;
    close(): Promise<void>;
}

// Factory functions
export function createServer(): RPCServer;
export function createClient(address: string): Promise<RPCClient>;


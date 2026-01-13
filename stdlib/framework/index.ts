// Standard Library: Runtime-aware Framework
// TypeScript definitions for the official framework

export interface Request {
    method: string;
    path: string;
    headers: Record<string, string>;
    body: Uint8Array | string;
    query: Record<string, string>;
    params: Record<string, string>;
}

export interface Response {
    status: number;
    headers: Record<string, string>;
    body: Uint8Array | string;
}

export interface Context {
    request: Request;
    response: Response;
    data: Record<string, any>;
    set(key: string, value: any): void;
    get(key: string): any;
}

export type Middleware = (ctx: Context, next: () => Promise<void> | void) => Promise<void> | void;
export type Handler = (ctx: Context) => Promise<void> | void;

export interface App {
    use(middleware: Middleware): void;
    get(path: string, handler: Handler): void;
    post(path: string, handler: Handler): void;
    put(path: string, handler: Handler): void;
    delete(path: string, handler: Handler): void;
    onStart(hook: () => Promise<void> | void): void;
    onStop(hook: () => Promise<void> | void): void;
    start(): Promise<void>;
    stop(): Promise<void>;
    listen(port: number, callback?: (err?: Error) => void): void;
}

// Factory function to create a new application
export function createApp(name?: string): App;


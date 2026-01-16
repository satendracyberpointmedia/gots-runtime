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

    json(data: any): void;
    text(text: string): void;
    html(html: string): void;
    setStatus(code: number): Response;
    setHeader(name: string, value: string): Response;
}

export interface Context {
    request: Request;
    response: Response;
    data: Record<string, any>;
    set(key: string, value: any): void;
    get(key: string): any;
    param(name: string): string | undefined;
    query(name: string): string | undefined;
    header(name: string): string | undefined;
}

export type Middleware = (ctx: Context, next: () => Promise<void> | void) => Promise<void> | void;
export type Handler = (ctx: Context) => Promise<void> | void;
export type ErrorHandler = (ctx: Context, error: Error) => Promise<void> | void;
export type NotFoundHandler = (ctx: Context) => Promise<void> | void;

export interface App {
    use(middleware: Middleware): App;
    get(path: string, handler: Handler): App;
    post(path: string, handler: Handler): App;
    put(path: string, handler: Handler): App;
    delete(path: string, handler: Handler): App;
    patch(path: string, handler: Handler): App;
    options(path: string, handler: Handler): App;
    head(path: string, handler: Handler): App;
    dynamic(method: string, path: string, handler: Handler): App;

    onStart(hook: () => Promise<void> | void): App;
    onStop(hook: () => Promise<void> | void): App;

    setErrorHandler(handler: ErrorHandler): App;
    setNotFoundHandler(handler: NotFoundHandler): App;

    start(): Promise<void>;
    stop(): Promise<void>;
    handle(ctx: Context): Promise<void>;
    listen(port: number, callback?: (err?: Error) => void): void;
}

// Factory function to create a new application
export function createApp(name?: string): App { throw new Error('Not implemented'); }

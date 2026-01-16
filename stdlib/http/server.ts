// Standard Library: HTTP Server & Client
// This is a TypeScript definition file for the HTTP server & client APIs

export interface Request {
    method: string;
    url: string;
    path: string;
    headers: Record<string, string>;
    body: Uint8Array;
    params: Record<string, string>;
    query: Record<string, string>;
    remoteAddr: string;
    requestId?: string;
}

export interface Response {
    status: number;
    headers: Record<string, string>;
    body: Uint8Array | string;

    json(data: any): Response;
    text(text: string): Response;
    html(html: string): Response;
    buffer(buf: Uint8Array): Response;

    setStatus(code: number): Response;
    setHeader(name: string, value: string): Response;
    addHeader(name: string, value: string): Response;
    removeHeader(name: string): Response;

    redirect(url: string, statusCode?: number): void;
    send(data: any): void;
}

export type Handler = (req: Request, res: Response) => Promise<void> | void;
export type Middleware = (req: Request, res: Response, next: () => Promise<void>) => Promise<void> | void;
export type ErrorHandler = (error: Error, req: Request, res: Response) => Promise<void> | void;

export interface ServerOptions {
    host?: string;
    port?: number;
    timeout?: number;
    maxHeaderSize?: number;
    maxBodySize?: number;
    readTimeout?: number;
    writeTimeout?: number;
    idleTimeout?: number;
    keepAlive?: boolean;
    keepAliveTimeout?: number;
}

export interface Server {
    use(middleware: Middleware): Server;
    get(path: string, handler: Handler): Server;
    post(path: string, handler: Handler): Server;
    put(path: string, handler: Handler): Server;
    delete(path: string, handler: Handler): Server;
    patch(path: string, handler: Handler): Server;
    options(path: string, handler: Handler): Server;
    head(path: string, handler: Handler): Server;

    listen(port: number, host?: string, callback?: (err?: Error) => void): void;
    close(callback?: (err?: Error) => void): void;

    setErrorHandler(handler: ErrorHandler): Server;
    getStats(): {
        requestsTotal: number;
        requestsActive: number;
        bytesIn: number;
        bytesOut: number;
    };
}

export interface ClientOptions {
    timeout?: number;
    maxRedirects?: number;
    followRedirects?: boolean;
    keepAlive?: boolean;
    keepAliveTimeout?: number;
}

export interface ClientRequest {
    method: string;
    url: string;
    headers?: Record<string, string>;
    body?: Uint8Array | string;
    query?: Record<string, string>;
    timeout?: number;
}

export interface ClientResponse {
    status: number;
    headers: Record<string, string>;
    body: Uint8Array;
    text(): string;
    json(): any;
}

export interface Client {
    get(url: string, options?: ClientOptions): Promise<ClientResponse>;
    post(url: string, body?: any, options?: ClientOptions): Promise<ClientResponse>;
    put(url: string, body?: any, options?: ClientOptions): Promise<ClientResponse>;
    delete(url: string, options?: ClientOptions): Promise<ClientResponse>;
    patch(url: string, body?: any, options?: ClientOptions): Promise<ClientResponse>;
    request(req: ClientRequest): Promise<ClientResponse>;
    close(): Promise<void>;
}

// Factory functions
export function createServer(options?: ServerOptions): Server { throw new Error('Not implemented'); }
export function createClient(options?: ClientOptions): Client { throw new Error('Not implemented'); }

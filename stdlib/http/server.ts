// Standard Library: HTTP Server
// This is a TypeScript definition file for the HTTP server API

export interface Request {
    method: string;
    url: string;
    headers: Record<string, string>;
    body: Uint8Array;
    params: Record<string, string>;
    query: Record<string, string>;
}

export interface Response {
    status: number;
    headers: Record<string, string>;
    body: Uint8Array | string;
}

export type Handler = (req: Request) => Promise<Response> | Response;

export interface Server {
    handle(path: string, handler: Handler): void;
    listen(addr: string, callback: (err?: Error) => void): void;
    shutdown(callback: (err?: Error) => void): void;
}


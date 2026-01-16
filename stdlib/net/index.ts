// Standard Library: Network
// TypeScript definitions for network operations

export const SOCK_STREAM: number;
export const SOCK_DGRAM: number;
export const AF_INET: number;
export const AF_INET6: number;
export const AF_UNIX: number;

export interface Conn {
    read(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    readSync(buffer: Uint8Array): number;
    write(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    writeSync(buffer: Uint8Array): number;
    close(callback: (err?: Error) => void): void;
    closeSync(): void;
    localAddr(): string;
    remoteAddr(): string;
    setDeadline(t: number, callback: (err?: Error) => void): void;
    setReadDeadline(t: number, callback: (err?: Error) => void): void;
    setWriteDeadline(t: number, callback: (err?: Error) => void): void;
    setNoDelay(noDelay: boolean, callback?: (err?: Error) => void): void;
    setKeepAlive(keepAlive: boolean, interval?: number, callback?: (err?: Error) => void): void;
    getRawConn(): any;
}

export interface Listener {
    accept(callback: (conn: Conn, err?: Error) => void): void;
    acceptSync(): Conn;
    close(callback: (err?: Error) => void): void;
    closeSync(): void;
    addr(): string;
}

export interface UDPConn {
    readFrom(buffer: Uint8Array, callback: (n: number, addr: string, err?: Error) => void): void;
    writeTo(buffer: Uint8Array, addr: string, callback: (n: number, err?: Error) => void): void;
    close(callback: (err?: Error) => void): void;
    localAddr(): string;
    setDeadline(t: number, callback: (err?: Error) => void): void;
    setReadDeadline(t: number, callback: (err?: Error) => void): void;
    setWriteDeadline(t: number, callback: (err?: Error) => void): void;
}

export interface Net {
    // TCP Operations
    dial(network: string, address: string, callback: (conn: Conn, err?: Error) => void): void;
    dialSync(network: string, address: string): Conn;
    dialTimeout(network: string, address: string, timeout: number, callback: (conn: Conn, err?: Error) => void): void;
    dialTimeoutSync(network: string, address: string, timeout: number): Conn;

    listen(network: string, address: string, callback: (listener: Listener, err?: Error) => void): void;
    listenSync(network: string, address: string): Listener;

    // UDP Operations
    dialUDP(network: string, localAddr: string, remoteAddr: string, callback: (conn: UDPConn, err?: Error) => void): void;
    listenUDP(network: string, address: string, callback: (conn: UDPConn, err?: Error) => void): void;

    // DNS Resolution
    resolveTCPAddr(network: string, address: string, callback: (addr: string, err?: Error) => void): void;
    resolveUDPAddr(network: string, address: string, callback: (addr: string, err?: Error) => void): void;
    resolveIPAddr(network: string, address: string, callback: (addr: string, err?: Error) => void): void;

    lookupIP(host: string, callback: (ips: string[], err?: Error) => void): void;
    lookupIPv4(host: string, callback: (ips: string[], err?: Error) => void): void;
    lookupIPv6(host: string, callback: (ips: string[], err?: Error) => void): void;
    lookupHost(host: string, callback: (addrs: string[], err?: Error) => void): void;
    lookupPort(network: string, service: string, callback: (port: number, err?: Error) => void): void;
    lookupCNAME(host: string, callback: (cname: string, err?: Error) => void): void;

    lookupIPSync(host: string): string[];
    lookupIPv4Sync(host: string): string[];
    lookupIPv6Sync(host: string): string[];
    lookupHostSync(host: string): string[];
    lookupPortSync(network: string, service: string): number;
    lookupCNAMESync(host: string): string;

    // Connection pooling
    createPool(network: string, address: string, options?: PoolOptions): ConnPool;
}

export interface PoolOptions {
    maxConnections?: number;
    maxIdleConnections?: number;
    maxConnLifetime?: number;
    idleTimeout?: number;
}

export interface ConnPool {
    get(callback: (conn: Conn, err?: Error) => void): void;
    getSync(): Conn;
    release(conn: Conn, err?: Error): void;
    closeAll(callback?: (err?: Error) => void): void;
    getStats(): {
        openConnections: number;
        idleConnections: number;
        waitingRequests: number;
    };
}

export const net: Net;

export function getNet(): Net { throw new Error('Not implemented'); }

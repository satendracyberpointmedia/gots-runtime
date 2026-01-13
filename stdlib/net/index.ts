// Standard Library: Network
// TypeScript definitions for network operations

export interface Conn {
    read(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    write(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    close(callback: (err?: Error) => void): void;
    localAddr(): string;
    remoteAddr(): string;
    setDeadline(t: number, callback: (err?: Error) => void): void;
    setReadDeadline(t: number, callback: (err?: Error) => void): void;
    setWriteDeadline(t: number, callback: (err?: Error) => void): void;
}

export interface Listener {
    accept(callback: (conn: Conn, err?: Error) => void): void;
    close(callback: (err?: Error) => void): void;
    addr(): string;
}

export interface Net {
    dial(network: string, address: string, callback: (conn: Conn, err?: Error) => void): void;
    dialTimeout(network: string, address: string, timeout: number, callback: (conn: Conn, err?: Error) => void): void;
    listen(network: string, address: string, callback: (listener: Listener, err?: Error) => void): void;
    resolveTCPAddr(network: string, address: string, callback: (addr: string, err?: Error) => void): void;
    resolveUDPAddr(network: string, address: string, callback: (addr: string, err?: Error) => void): void;
    lookupIP(host: string, callback: (ips: string[], err?: Error) => void): void;
    lookupHost(host: string, callback: (addrs: string[], err?: Error) => void): void;
}


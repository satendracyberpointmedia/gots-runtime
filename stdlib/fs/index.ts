// Standard Library: File System
// TypeScript definitions for file system operations

export interface FileHandle {
    read(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    write(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    close(callback: (err?: Error) => void): void;
    seek(offset: number, whence: number, callback: (pos: number, err?: Error) => void): void;
}

export interface FS {
    readFile(path: string, callback: (data: Uint8Array, err?: Error) => void): void;
    writeFile(path: string, data: Uint8Array, perm: number, callback: (err?: Error) => void): void;
    readDir(path: string, callback: (entries: DirEntry[], err?: Error) => void): void;
    stat(path: string, callback: (info: FileInfo, err?: Error) => void): void;
    mkdir(path: string, perm: number, callback: (err?: Error) => void): void;
    mkdirAll(path: string, perm: number, callback: (err?: Error) => void): void;
    remove(path: string, callback: (err?: Error) => void): void;
    removeAll(path: string, callback: (err?: Error) => void): void;
    rename(oldpath: string, newpath: string, callback: (err?: Error) => void): void;
    exists(path: string, callback: (exists: boolean, err?: Error) => void): void;
    open(path: string, flag: number, perm: number, callback: (handle: FileHandle, err?: Error) => void): void;
    
    // Synchronous versions
    readFileSync(path: string): Uint8Array;
    writeFileSync(path: string, data: Uint8Array, perm: number): void;
    statSync(path: string): FileInfo;
    
    // Path utilities
    abs(path: string): string;
    join(...paths: string[]): string;
    base(path: string): string;
    dir(path: string): string;
    ext(path: string): string;
}

export interface DirEntry {
    name: string;
    isDir: boolean;
}

export interface FileInfo {
    name: string;
    size: number;
    mode: number;
    modTime: number;
    isDir: boolean;
}


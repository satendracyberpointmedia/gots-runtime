// Standard Library: File System
// TypeScript definitions for file system operations

export const O_RDONLY: number;
export const O_WRONLY: number;
export const O_RDWR: number;
export const O_APPEND: number;
export const O_CREATE: number;
export const O_EXCL: number;
export const O_SYNC: number;
export const O_TRUNC: number;

export interface FileHandle {
    read(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    readSync(buffer: Uint8Array): number;
    write(buffer: Uint8Array, callback: (n: number, err?: Error) => void): void;
    writeSync(buffer: Uint8Array): number;
    close(callback: (err?: Error) => void): void;
    closeSync(): void;
    seek(offset: number, whence: number, callback: (pos: number, err?: Error) => void): void;
    seekSync(offset: number, whence: number): number;
    stat(callback: (info: FileInfo, err?: Error) => void): void;
    statSync(): FileInfo;
    truncate(size: number, callback: (err?: Error) => void): void;
    truncateSync(size: number): void;
    chmod(mode: number, callback: (err?: Error) => void): void;
    chmodSync(mode: number): void;
}

export interface FS {
    // Async file operations
    readFile(path: string, callback: (data: Uint8Array, err?: Error) => void): void;
    readFile(path: string, encoding: 'utf8' | 'ascii' | 'base64', callback: (data: string, err?: Error) => void): void;

    writeFile(path: string, data: Uint8Array | string, perm: number, callback: (err?: Error) => void): void;
    appendFile(path: string, data: Uint8Array | string, callback: (err?: Error) => void): void;

    readDir(path: string, callback: (entries: DirEntry[], err?: Error) => void): void;

    stat(path: string, callback: (info: FileInfo, err?: Error) => void): void;
    lstat(path: string, callback: (info: FileInfo, err?: Error) => void): void;

    mkdir(path: string, perm: number, callback: (err?: Error) => void): void;
    mkdirAll(path: string, perm: number, callback: (err?: Error) => void): void;

    remove(path: string, callback: (err?: Error) => void): void;
    removeAll(path: string, callback: (err?: Error) => void): void;

    rename(oldpath: string, newpath: string, callback: (err?: Error) => void): void;
    symlink(oldname: string, newname: string, callback: (err?: Error) => void): void;
    readlink(path: string, callback: (target: string, err?: Error) => void): void;

    chmod(path: string, mode: number, callback: (err?: Error) => void): void;
    chown(path: string, uid: number, gid: number, callback: (err?: Error) => void): void;

    copy(src: string, dst: string, callback: (err?: Error) => void): void;
    copyRecursive(src: string, dst: string, callback: (err?: Error) => void): void;

    exists(path: string, callback: (exists: boolean, err?: Error) => void): void;
    isDir(path: string, callback: (isDir: boolean, err?: Error) => void): void;
    isFile(path: string, callback: (isFile: boolean, err?: Error) => void): void;

    open(path: string, flag: number, perm: number, callback: (handle: FileHandle, err?: Error) => void): void;

    // Sync file operations
    readFileSync(path: string, encoding?: 'utf8' | 'ascii' | 'base64'): Uint8Array | string;
    writeFileSync(path: string, data: Uint8Array | string, perm: number): void;
    appendFileSync(path: string, data: Uint8Array | string): void;

    readDirSync(path: string): DirEntry[];
    statSync(path: string): FileInfo;
    lstatSync(path: string): FileInfo;

    mkdirSync(path: string, perm: number): void;
    mkdirAllSync(path: string, perm: number): void;

    removeSync(path: string): void;
    removeAllSync(path: string): void;

    renameSync(oldpath: string, newpath: string): void;
    symlinkSync(oldname: string, newname: string): void;
    readlinkSync(path: string): string;

    chmodSync(path: string, mode: number): void;
    chownSync(path: string, uid: number, gid: number): void;

    copySync(src: string, dst: string): void;
    copyRecursiveSync(src: string, dst: string): void;

    existsSync(path: string): boolean;
    isDirSync(path: string): boolean;
    isFileSync(path: string): boolean;

    openSync(path: string, flag: number, perm: number): FileHandle;

    // Path utilities
    abs(path: string): string;
    join(...paths: string[]): string;
    base(path: string): string;
    dir(path: string): string;
    ext(path: string): string;
    relative(from: string, to: string): string;
    resolve(...paths: string[]): string;
    normalize(path: string): string;

    // Glob operations
    glob(pattern: string, callback: (matches: string[], err?: Error) => void): void;
    globSync(pattern: string): string[];

    // Watch operations
    watch(path: string, callback: (event: string, filename: string, err?: Error) => void): WatchHandle;
}

export interface DirEntry {
    name: string;
    isDir: boolean;
    size?: number;
    modTime?: number;
    isSymlink?: boolean;
}

export interface FileInfo {
    name: string;
    size: number;
    mode: number;
    modTime: number;
    isDir: boolean;
    isSymlink?: boolean;
    uid?: number;
    gid?: number;
}

export interface WatchHandle {
    close(callback?: (err?: Error) => void): void;
}

export const fs: FS;

export function getFS(): FS { throw new Error('Not implemented'); }

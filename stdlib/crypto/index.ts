// Standard Library: Crypto
// TypeScript definitions for cryptographic operations

export interface Crypto {
    md5(data: Uint8Array): string;
    sha1(data: Uint8Array): string;
    sha256(data: Uint8Array): string;
    sha512(data: Uint8Array): string;
    randomBytes(n: number): Uint8Array;
    randomHex(n: number): string;
    randomUUID(): string;
}


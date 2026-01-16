// Standard Library: Crypto
// TypeScript definitions for cryptographic operations

export interface Crypto {
    // Hashing
    md5(data: Uint8Array | string): string;
    sha1(data: Uint8Array | string): string;
    sha256(data: Uint8Array | string): string;
    sha512(data: Uint8Array | string): string;
    sha3_256(data: Uint8Array | string): string;
    sha3_512(data: Uint8Array | string): string;
    blake2b(data: Uint8Array | string, size?: number): string;
    blake2s(data: Uint8Array | string, size?: number): string;

    // Random
    randomBytes(n: number): Uint8Array;
    randomHex(n: number): string;
    randomUUID(): string;
    randomInt(min: number, max: number): number;
    randomString(length: number, charset?: string): string;

    // HMAC
    hmac(algorithm: string, key: Uint8Array | string, data: Uint8Array | string): string;
    hmacSha256(key: Uint8Array | string, data: Uint8Array | string): string;
    hmacSha512(key: Uint8Array | string, data: Uint8Array | string): string;

    // Encryption/Decryption
    encrypt(algorithm: string, key: Uint8Array, data: Uint8Array, iv?: Uint8Array): Uint8Array;
    decrypt(algorithm: string, key: Uint8Array, data: Uint8Array, iv?: Uint8Array): Uint8Array;

    // AES
    aesGcmEncrypt(key: Uint8Array, data: Uint8Array, aad?: Uint8Array): { ciphertext: Uint8Array, nonce: Uint8Array, tag: Uint8Array };
    aesGcmDecrypt(key: Uint8Array, ciphertext: Uint8Array, nonce: Uint8Array, tag: Uint8Array, aad?: Uint8Array): Uint8Array;

    // RSA
    generateRSAKeyPair(keySize?: number): { publicKey: string, privateKey: string };
    rsaEncrypt(publicKey: string, data: Uint8Array): Uint8Array;
    rsaDecrypt(privateKey: string, data: Uint8Array): Uint8Array;
    rsaSign(privateKey: string, data: Uint8Array, algorithm?: string): Uint8Array;
    rsaVerify(publicKey: string, data: Uint8Array, signature: Uint8Array, algorithm?: string): boolean;

    // ECDSA
    generateECDSAKeyPair(curve?: string): { publicKey: string, privateKey: string };
    ecdsaSign(privateKey: string, data: Uint8Array): { r: Uint8Array, s: Uint8Array };
    ecdsaVerify(publicKey: string, data: Uint8Array, r: Uint8Array, s: Uint8Array): boolean;

    // Hashing stream
    createHash(algorithm: string): HashStream;
    createHmac(algorithm: string, key: Uint8Array | string): HmacStream;

    // Key derivation
    pbkdf2(password: string, salt: Uint8Array, iterations: number, length: number, digest?: string): Uint8Array;
    scrypt(password: string, salt: Uint8Array, n: number, r: number, p: number, length: number): Uint8Array;
    argon2(password: string, salt: Uint8Array, time: number, memory: number, parallelism: number, length: number): Uint8Array;

    // Utility
    toHex(data: Uint8Array): string;
    fromHex(hex: string): Uint8Array;
    toBase64(data: Uint8Array): string;
    fromBase64(base64: string): Uint8Array;
    timingSafeEqual(a: Uint8Array, b: Uint8Array): boolean;
}

export interface HashStream {
    update(data: Uint8Array | string): HashStream;
    digest(encoding?: 'hex' | 'base64' | 'binary'): Uint8Array | string;
    reset(): HashStream;
}

export interface HmacStream {
    update(data: Uint8Array | string): HmacStream;
    digest(encoding?: 'hex' | 'base64' | 'binary'): Uint8Array | string;
    reset(): HmacStream;
}

// Singleton instance
export declare const crypto: Crypto;

// Factory function
export function getCrypto(): Crypto { throw new Error('Not implemented'); }

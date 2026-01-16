// Cryptography Example
// Run with: gots run crypto.ts

import { getCrypto } from "../stdlib/crypto/index";

async function cryptoExamples(): Promise<void> {
    const crypto = getCrypto();

    console.log("=== Hashing Examples ===");

    // Hash examples
    const data = "Hello, GoTS Runtime!";
    console.log("Data:", data);

    const md5 = crypto.md5(data);
    console.log("MD5:", md5);

    const sha256 = crypto.sha256(data);
    console.log("SHA256:", sha256);

    const sha512 = crypto.sha512(data);
    console.log("SHA512:", sha512);

    console.log("\n=== Random Generation ===");

    const randomBytes = crypto.randomBytes(16);
    console.log("Random bytes:", randomBytes);

    const randomHex = crypto.randomHex(8);
    console.log("Random hex:", randomHex);

    const randomUUID = crypto.randomUUID();
    console.log("Random UUID:", randomUUID);

    const randomString = crypto.randomString(20, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789");
    console.log("Random string:", randomString);

    console.log("\n=== HMAC Examples ===");

    const key = "secret_key";
    const hmacResult = crypto.hmacSha256(key, data);
    console.log("HMAC-SHA256:", hmacResult);

    console.log("\n=== Utility Functions ===");

    const hex = crypto.toHex(new TextEncoder().encode("test"));
    console.log("Text to Hex:", hex);

    const base64 = crypto.toBase64(new TextEncoder().encode("test"));
    console.log("Text to Base64:", base64);

    console.log("\nCrypto examples completed!");
}

cryptoExamples().catch(console.error);

export { cryptoExamples };

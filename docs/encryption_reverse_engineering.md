## Reverse Engineering the Liquid Telecom Router Encryption

This document details the analysis and implementation of the encryption protocols used by the Liquid Telecom router (specifically the model targeted by `tool-iquid`). The router enforces client-side encryption for sensitive actions like **Login** and **Reboot**, requiring a custom client implementation to interact with it programmatically.

### 1. Overview of the Security Mechanism

The router does not use standard HTTP Basic Auth or simple form POSTs. Instead, it employs a dual-layer encryption scheme using **RSA** (for key exchange) and **AES** (for payload encryption). This mimics a TLS-like handshake but implemented entirely in JavaScript within the browser.

#### The Handshake Flow
1.  **Client Request (GET)**: The client requests a page (e.g., `index.html` or `reboot.cgi`).
2.  **Server Response**: The server embeds a unique **Public Key (RSA)**, a **Nonce**, and a **CSRF Token** directly into the HTML/JavaScript of the page.
    *   *Example:* `var pubkey = '-----BEGIN PUBLIC KEY-----...';`
3.  **Client Encryption**:
    *   Generates ephemeral **AES Key** and **IV** (Initialization Vector).
    *   Encrypts the payload (credentials, tokens) using **AES**.
    *   Encrypts the AES Key + IV using the server's **RSA Public Key**.
4.  **Submission (POST)**: The client attempts the action by sending two encrypted blobs:
    *   `ct`: Cipher Text (The AES-encrypted payload).
    *   `ck`: Cipher Key (The RSA-encrypted AES key/IV).

### 2. Cryptographic Primitives

Through analysis of the router's `crypto_page.js` and `jsencrypt.min.js`, we identified the following standard algorithms:

#### RSA (Rivest–Shamir–Adleman)
*   **Purpose**: Securely transmitting the ephemeral AES keys to the server.
*   **Library**: The router uses `JSEncrypt`.
*   **Padding**: `PKCS1v1.5`.
*   **Format**: The public key is provided in standard PEM format (Base64 encoded DER with headers).

#### AES (Advanced Encryption Standard)
*   **Purpose**: Encrypting the actual form data (username, password, etc.).
*   **Library**: The router uses `sjcl` (Stanford Javascript Crypto Library).
*   **Mode**: `CBC` (Cipher Block Chaining).
*   **Key Size**: 128-bit (16 bytes).
*   **Padding**: `PKCS7` (standard for block ciphers).

#### Base64URL Encoding (Custom)
A critical discovery was the router's specific flavor of Base64 encoding. Standard Base64 uses `+`, `/`, and `=`. The router's `sjcl.codec.base64url` implementation uses a URL-safe variant with specific replacements:

*   `+` becomes `-`
*   `/` becomes `_`
*   `=` (padding) becomes `.`

*Implementation in `pkg/liquid/crypto.go`:*
```go
func Base64UrlEscape(b64 string) string {
    b64 = strings.ReplaceAll(b64, "+", "-")
    b64 = strings.ReplaceAll(b64, "/", "_")
    b64 = strings.ReplaceAll(b64, "=", ".")
    return b64
}
```

### 3. Reverse Engineering Process

#### Phase A: Static Analysis of JavaScript
We started by examining the `login.cgi` page source and identified `encrypt_post_data` in `crypto_page.js`.

***:
    ```javascript
    var encrypt_post_data =   **Logic Found function(pubkey, plaintext) {
       var crypt = new JSEncrypt();
       crypt.setPublicKey(pubkey);
       var p = encrypt(pubkey, plaintext); // Calls internal AES logic
       return  'encrypted=1&ct=' + p.ct + '&ck=' + p.ck;
    };
    ```
*   This confirmed we needed to construct a POST body with `encrypted=1`, `ct`, and `ck`.

#### Phase B: Payload Reconstruction
The next challenge was determining *what* specifically was being encrypted (the `plaintext`).

*   **Login Payload**:
    It was not just `username=...&password=...`. Debugging the `login.cgi` behavior revealed a specific structure constructed in JavaScript:
    ```javascript
    // Reconstructed format
    let data = '&username=' + user + '&password=' + urlEncodedPass +
               '&csrf_token=' + token + '&nonce=' + nonce +
               '&enckey=' + base64(dec_key) + '&enciv=' + base64(dec_iv);
    ```
    *Note: The payload strangely includes *decryption* keys (`enckey`/`enciv`) which suggests the server might use these to encrypt its response back to the client, allowing for a bidirectional encrypted tunnel.*

*   **Reboot Payload**:
    The reboot logic was harder to trace. We implemented a temporary `-dump` flag (since removed) to extract `reboot.cgi` and its assets for inspection.
    *   **Finding**: `reboot.cgi` does *not* use the global login variables. It has its own isolated `pubkey` and `csrf_token` in the HTML.
    *   **Structure**: The payload for reboot is simpler: `data&csrf_token=<token>`.

#### Phase C: Network Traffic Analysis
We verified our assumptions by inspecting actual traffic (via browser DevTools and `tcpdump`).
*   **Observation**: The `ck` parameter (RSA encrypted key) changes with every request (confirming ephemeral keys).
*   **Observation**: The `password` field in the plaintext must be `encodeURIComponent`'d (URL encoded) *before* encryption.

### 4. Implementation Details (`tool-iquid`)

The Golang implementation in `pkg/liquid` replicates this entire stack:

#### `crypto.go`
*   `EncryptRSA`: Uses `crypto/rsa` with `EncryptPKCS1v15`.
*   `EncryptAES`: Uses `crypto/aes` and `crypto/cipher` (CBC mode) with logic to properly pad data (PKCS7).
*   `Base64UrlEscape`: Matches the router's custom encoding.

#### `client.go`
*   **Login Flow**:
    1.  GET `/` -> Regex parse `nonce`, `token`, `pubkey`.
    2.  Generate 16-byte `AES Key` and `IV`.
    3.  Construct payload string.
    4.  Encrypt -> POST to `/login.cgi`.
*   **Reboot Flow**:
    1.  GET `/reboot.cgi` -> Regex parse page-specific `token` and `pubkey`.
    2.  Same AES/RSA encryption process.
    3.  POST to `/reboot.cgi?reboot`.

### 5. Security Implications
While this scheme is more complex than Basic Auth, it is technically **Client-Side Encryption**, which does not replace TLS (HTTPS).
*   **Vulnerability**: The public key is sent in plaintext (if HTTP). An attacker on the wire can replace the key (MITM) or simply capture the `ck` and `ct` to replay them within the token's validity window.
*   **Session Binding**: The use of `nonce` and `csrf_token` limits replay attacks, but the security ultimately relies on the transport layer (which is plain HTTP in this router's default configuration).

---
*Document generated by AntiGravity Agent.*

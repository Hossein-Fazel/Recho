/**
 * Generates a RFC 4122 version 4 UUID for websocket request ids.
 *
 * `crypto.randomUUID()` is only exposed in secure contexts, so it is missing
 * when the app is served over plain HTTP — which is what happens when a phone
 * opens the app at `http://<lan-ip>:5173`. Reading it off `crypto` there yields
 * `undefined` and calling it throws, silently killing the send/edit/delete
 * handlers. `crypto.getRandomValues()` has no such restriction, so use it as
 * the fallback.
 *
 * The server decodes `request_id` into a `uuid.UUID`, so every branch has to
 * produce a well-formed UUID string.
 */
export function newRequestId(): string {
  const webCrypto = globalThis.crypto

  if (webCrypto?.randomUUID) {
    return webCrypto.randomUUID()
  }

  const bytes = new Uint8Array(16)

  if (webCrypto?.getRandomValues) {
    webCrypto.getRandomValues(bytes)
  } else {
    for (let i = 0; i < bytes.length; i += 1) {
      bytes[i] = Math.floor(Math.random() * 256)
    }
  }

  // Version 4, variant 1.
  bytes[6] = (bytes[6] & 0x0f) | 0x40
  bytes[8] = (bytes[8] & 0x3f) | 0x80

  const hex = Array.from(bytes, (byte) =>
    byte.toString(16).padStart(2, '0'),
  ).join('')

  return [
    hex.slice(0, 8),
    hex.slice(8, 12),
    hex.slice(12, 16),
    hex.slice(16, 20),
    hex.slice(20),
  ].join('-')
}

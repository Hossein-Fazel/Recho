/** Encode a cursor the Go backend can decode (json + raw URL-safe base64). */
function encodeGoCursor(fields: Record<string, unknown>): string {
  const json = JSON.stringify(fields)
  const bytes = new TextEncoder().encode(json)
  let binary = ''
  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

/** First page: zero time/id so the SQL cursor clause is treated as unset. */
export const firstConvCursor = encodeGoCursor({
  Date: '0001-01-01T00:00:00Z',
  ID: '00000000-0000-0000-0000-000000000000',
})

export const firstMessageCursor = encodeGoCursor({
  Date: '0001-01-01T00:00:00Z',
  ID: 0,
})

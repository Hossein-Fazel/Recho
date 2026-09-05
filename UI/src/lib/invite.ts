/** Invite links look like `https://host/join/ABC123XYZ0`. */
const JOIN_PREFIX = '/join/'

export function inviteLink(code: string): string {
  return `${window.location.origin}${JOIN_PREFIX}${encodeURIComponent(code)}`
}

/** Reads the invite code out of a pathname, or '' when it isn't a join link. */
export function inviteCodeFromPath(pathname: string): string {
  if (!pathname.startsWith(JOIN_PREFIX)) return ''

  const raw = pathname.slice(JOIN_PREFIX.length).split('/')[0]

  try {
    return decodeURIComponent(raw).trim()
  } catch {
    return raw.trim()
  }
}

/** Drops the invite path from the URL bar without reloading the page. */
export function clearInvitePath() {
  window.history.replaceState(null, '', '/')
}

import { inviteCodeFromPath } from './invite'

export type TextPart = {
  kind: 'text'
  value: string
}

export type LinkPart = {
  kind: 'link'
  /** The text exactly as it was typed, so messages still read as written. */
  value: string
  href: string
  /** Non-empty when the link is an invite for this deployment. */
  inviteCode: string
}

export type MessagePart = TextPart | LinkPart

/**
 * Matches an explicit `http(s)://` URL or a `www.` host.
 *
 * Bare domains are deliberately not matched: `main.go`, `index.css` and `v1.2`
 * show up in chat far more often than someone typing a URL without a scheme,
 * and linkifying those is worse than missing the occasional real link.
 */
const urlPattern = /(?:https?:\/\/|www\.)[^\s]+/gi

const trailingPunctuation = new Set([
  '.', ',', ';', ':', '!', '?', '"', "'", '`', '’', '”', '»',
])

const closers: Record<string, string> = {
  ')': '(',
  ']': '[',
  '}': '{',
}

function occurrences(text: string, char: string): number {
  let total = 0

  for (const candidate of text) {
    if (candidate === char) total += 1
  }

  return total
}

/**
 * Drops punctuation that belongs to the sentence rather than the URL, so
 * `see https://recho.dev.` links to `https://recho.dev` while
 * `https://en.wikipedia.org/wiki/Go_(language)` keeps its closing paren.
 */
function trimTrailing(url: string): string {
  let end = url.length

  while (end > 0) {
    const char = url[end - 1]

    if (trailingPunctuation.has(char)) {
      end -= 1
      continue
    }

    const opener = closers[char]

    if (opener) {
      const slice = url.slice(0, end)

      // Unbalanced closer: it wrapped the URL instead of being part of it.
      if (occurrences(slice, char) > occurrences(slice, opener)) {
        end -= 1
        continue
      }
    }

    break
  }

  return url.slice(0, end)
}

/**
 * Builds a safe href, or '' when the candidate isn't a usable web URL.
 *
 * `urlPattern` already makes `javascript:` and `data:` unmatchable, but the
 * protocol is re-checked here so the guarantee lives next to the href instead
 * of only in the regex.
 */
function resolve(raw: string): URL | null {
  const hasScheme = /^https?:\/\//i.test(raw)

  try {
    const url = new URL(hasScheme ? raw : `https://${raw}`)

    if (url.protocol !== 'http:' && url.protocol !== 'https:') return null
    if (!url.hostname) return null

    // `www.` shorthand always carries a dot; a scheme-less host without one is
    // leftover prose, not a domain.
    if (!hasScheme && !url.hostname.includes('.')) return null

    return url
  } catch {
    return null
  }
}

/** Splits message text into plain runs and links, preserving every character. */
export function splitLinks(text: string): MessagePart[] {
  const parts: MessagePart[] = []
  let cursor = 0

  urlPattern.lastIndex = 0

  for (
    let match = urlPattern.exec(text);
    match;
    match = urlPattern.exec(text)
  ) {
    const value = trimTrailing(match[0])

    // Re-scan from the end of the trimmed URL: whatever was stripped is
    // ordinary text. `value` always keeps its `http`/`www` prefix, so this
    // still advances and cannot loop forever.
    urlPattern.lastIndex = match.index + value.length

    const url = resolve(value)
    if (!url) continue

    if (match.index > cursor) {
      parts.push({ kind: 'text', value: text.slice(cursor, match.index) })
    }

    parts.push({
      kind: 'link',
      value,
      href: url.href,
      inviteCode:
        url.origin === window.location.origin
          ? inviteCodeFromPath(url.pathname)
          : '',
    })

    cursor = match.index + value.length
  }

  if (cursor < text.length) {
    parts.push({ kind: 'text', value: text.slice(cursor) })
  }

  return parts
}

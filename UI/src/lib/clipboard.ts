/**
 * Copies text to the clipboard.
 *
 * `navigator.clipboard` only exists in secure contexts, so it is missing when
 * the app is opened over plain HTTP — which is exactly what happens when a
 * phone visits the dev server at `http://<lan-ip>:5173`. Reading `.writeText`
 * off `undefined` there throws synchronously and takes the caller's click
 * handler down with it, so guard the API and fall back to the legacy
 * `execCommand` path.
 *
 * Resolves to whether the text made it to the clipboard.
 */
export async function copyText(value: string): Promise<boolean> {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      return true
    } catch {
      // Denied permission or a non-focused document; try the fallback below.
    }
  }

  return legacyCopy(value)
}

function legacyCopy(value: string): boolean {
  const area = document.createElement('textarea')

  area.value = value
  area.setAttribute('readonly', '')
  // Keep it out of sight and stop iOS from scrolling to / zooming into it.
  area.style.position = 'fixed'
  area.style.top = '-1000px'
  area.style.opacity = '0'

  document.body.append(area)

  const selection = document.getSelection()
  const previous = selection && selection.rangeCount > 0
    ? selection.getRangeAt(0)
    : null

  try {
    area.select()
    area.setSelectionRange(0, value.length)

    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    area.remove()

    if (previous && selection) {
      selection.removeAllRanges()
      selection.addRange(previous)
    }
  }
}

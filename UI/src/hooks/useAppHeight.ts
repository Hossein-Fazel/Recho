import { useEffect } from 'react'

/**
 * Keeps `--app-height` on <html> in sync with the visual viewport.
 *
 * Full-height panes use this instead of `100dvh` directly. `interactive-widget`
 * in the viewport meta already shrinks the layout viewport when the soft
 * keyboard opens on Chrome, but Safari ignores that hint and only shrinks the
 * *visual* viewport — leaving the composer behind the keyboard. Measuring
 * `visualViewport` covers both.
 *
 * The CSS variable falls back to `100dvh`, so the layout is correct before this
 * ever runs and on browsers without `visualViewport`.
 */
export function useAppHeight() {
  useEffect(() => {
    const viewport = window.visualViewport
    const root = document.documentElement

    function apply() {
      const height = viewport?.height ?? window.innerHeight
      root.style.setProperty('--app-height', `${Math.round(height)}px`)
    }

    apply()

    if (!viewport) {
      window.addEventListener('resize', apply)
      window.addEventListener('orientationchange', apply)

      return () => {
        window.removeEventListener('resize', apply)
        window.removeEventListener('orientationchange', apply)
        root.style.removeProperty('--app-height')
      }
    }

    viewport.addEventListener('resize', apply)

    return () => {
      viewport.removeEventListener('resize', apply)
      root.style.removeProperty('--app-height')
    }
  }, [])
}

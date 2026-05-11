import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'

export type SubtitleStatus = 'idle' | 'loading' | 'none' | 'available'

export function useSubtitleTracks(
  videoRef: React.RefObject<HTMLVideoElement | null>,
  movieId: number,
  isStreaming: boolean,
): SubtitleStatus {
  const { t, i18n } = useTranslation()
  const blobUrlsRef = useRef<string[]>([])
  const [status, setStatus] = useState<SubtitleStatus>('idle')

  useEffect(() => {
    if (!isStreaming) {
      setStatus('idle')
      return
    }

    const userLang = i18n.language.slice(0, 2)
    let cancelled = false
    setStatus('loading')

    void (async () => {
      // 1. Fetch the list of cached languages. If empty, stop — no further requests.
      let available: string[] = []
      try {
        const res = await fetch(`/api/v1/movies/${movieId}/subtitles`, { credentials: 'include' })
        if (!res.ok || cancelled) return
        const json = await res.json()
        available = (json?.data ?? json)?.languages ?? []
      } catch {
        if (!cancelled) setStatus('none')
        return
      }

      if (cancelled) return

      if (available.length === 0) {
        setStatus('none')
        return
      }

      // 2. Wait for the video element to be in the DOM.
      const video = videoRef.current
      if (!video || cancelled) {
        setStatus('none')
        return
      }

      // 3. Load each available language as a blob-URL track.
      for (const lang of available) {
        if (cancelled) break
        try {
          const res = await fetch(`/api/v1/movies/${movieId}/subtitles/${lang}`, { credentials: 'include' })
          if (!res.ok || cancelled) continue
          const blob = await res.blob()
          if (cancelled) continue

          const url = URL.createObjectURL(blob)
          blobUrlsRef.current.push(url)

          const el = document.createElement('track')
          el.kind = 'subtitles'
          el.srclang = lang
          el.label = t(`movie.subtitle_${lang}`, { defaultValue: lang.toUpperCase() })
          el.src = url
          if (lang === userLang) el.default = true
          video.appendChild(el)
        } catch { /* not available for this language */ }
      }

      if (!cancelled) setStatus('available')
    })()

    return () => {
      cancelled = true
      const video = videoRef.current
      if (video) Array.from(video.querySelectorAll('track')).forEach(el => el.remove())
      blobUrlsRef.current.forEach(url => URL.revokeObjectURL(url))
      blobUrlsRef.current = []
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isStreaming, movieId, i18n.language, t])

  return status
}

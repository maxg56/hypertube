import { useEffect, useRef } from 'react'

const SAVE_INTERVAL_MS = 5000

export function useProgressSync(
  videoRef: React.RefObject<HTMLVideoElement | null>,
  movieId: number,
  isStreaming: boolean,
) {
  const saveRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Restore saved position once when streaming starts.
  useEffect(() => {
    if (!isStreaming) return
    let cancelled = false

    const timer = setTimeout(async () => {
      const video = videoRef.current
      if (!video || cancelled) return
      try {
        const res = await fetch(`/api/v1/movies/${movieId}/progress`, { credentials: 'include' })
        if (!res.ok || cancelled) return
        const json = await res.json()
        const sec: number = (json.data ?? json).progress_sec ?? 0
        if (sec > 0 && videoRef.current) videoRef.current.currentTime = sec
      } catch { /* start from beginning */ }
    }, 300)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isStreaming, movieId])

  // Save position every SAVE_INTERVAL_MS with PUT only — no GET here.
  useEffect(() => {
    if (!isStreaming) return

    saveRef.current = setInterval(async () => {
      const video = videoRef.current
      if (!video || video.paused || video.ended) return
      const sec = Math.floor(video.currentTime)
      if (sec <= 0) return
      try {
        await fetch(`/api/v1/movies/${movieId}/progress`, {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ progress_sec: sec }),
        })
      } catch { /* ignore transient errors */ }
    }, SAVE_INTERVAL_MS)

    return () => {
      if (saveRef.current) {
        clearInterval(saveRef.current)
        saveRef.current = null
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isStreaming, movieId])
}

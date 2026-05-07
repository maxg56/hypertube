'use client'

import { useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'

interface Comment {
  id: number
  user_id: number
  username: string
  avatar_url: string
  content: string
  created_at: string
}

interface CommentSectionProps {
  movieId: number
  initialComments: Comment[]
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch {
    return iso
  }
}

export function CommentSection({ movieId, initialComments }: CommentSectionProps) {
  const { t } = useTranslation()
  const [comments, setComments] = useState<Comment[]>(initialComments)
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return
    setSubmitting(true)
    setError(null)

    try {
      const res = await fetch(`/api/v1/comments/${movieId}`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: content.trim() }),
      })
      if (!res.ok) throw new Error(`${res.status}`)
      const json = await res.json()
      const newComment: Comment = json.data ?? json
      setComments(prev => [newComment, ...prev])
      setContent('')
    } catch {
      setError(t('movie.comment_error'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (commentId: number) => {
    try {
      const res = await fetch(`/api/v1/comments/${commentId}`, {
        method: 'DELETE',
        credentials: 'include',
      })
      if (res.ok) {
        setComments(prev => prev.filter(c => c.id !== commentId))
      }
    } catch {
      // silently ignore
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">{t('movie.comments_title')}</h2>

      <form onSubmit={handleSubmit} className="flex flex-col gap-2">
        <textarea
          ref={textareaRef}
          value={content}
          onChange={e => setContent(e.target.value)}
          placeholder={t('movie.comment_placeholder')}
          rows={3}
          maxLength={2000}
          className="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm resize-none focus:outline-none focus:ring-1 focus:ring-sidebar-primary"
        />
        {error && <p className="text-xs text-destructive">{error}</p>}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={submitting || !content.trim()}
            className="px-4 py-1.5 rounded-md bg-sidebar-primary text-white text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {submitting ? t('movie.comment_submitting') : t('movie.comment_submit')}
          </button>
        </div>
      </form>

      <div className="flex flex-col gap-4">
        {comments.length === 0 && (
          <p className="text-sm text-muted-foreground">{t('movie.no_comments')}</p>
        )}
        {comments.map(comment => (
          <div key={comment.id} className="flex gap-3">
            <div className="shrink-0 w-9 h-9 rounded-full overflow-hidden bg-muted border border-border">
              {comment.avatar_url ? (
                <img src={comment.avatar_url} alt={comment.username} className="w-full h-full object-cover" />
              ) : (
                <span className="w-full h-full flex items-center justify-center text-sm font-bold text-muted-foreground">
                  {comment.username?.charAt(0).toUpperCase()}
                </span>
              )}
            </div>
            <div className="flex flex-col gap-1 min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{comment.username}</span>
                <span className="text-xs text-muted-foreground">{formatDate(comment.created_at)}</span>
              </div>
              <p className="text-sm text-muted-foreground break-words">{comment.content}</p>
            </div>
            <button
              onClick={() => handleDelete(comment.id)}
              className="shrink-0 text-xs text-muted-foreground hover:text-destructive transition-colors self-start mt-0.5"
              aria-label={t('movie.comment_delete')}
            >
              ✕
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}

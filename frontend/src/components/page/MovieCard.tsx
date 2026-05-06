'use client'

import { Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Movie } from '@/hooks/useMovies'

interface MovieCardProps {
  movie: Movie
  watched: boolean
  onToggle: () => void
}

export function MovieCard({ movie, watched, onToggle }: MovieCardProps) {
  return (
    <div className="group cursor-pointer flex flex-col" onClick={onToggle}>
      <div className={cn('card-glow bg-card border rounded-lg overflow-hidden flex flex-col flex-1', watched && 'opacity-50')}>
        <div className="bg-muted w-full aspect-[2/3] overflow-hidden flex items-center justify-center">
          {movie.poster_url ? (
            <img
              src={movie.poster_url}
              alt={movie.title}
              className={cn('w-full h-full object-cover group-hover:scale-105 transition-transform', watched && 'grayscale')}
            />
          ) : (
            <span className="text-muted-foreground text-center text-sm px-2">{movie.title}</span>
          )}
        </div>
        <div className="p-3 flex flex-col gap-1 flex-shrink-0">
          <div className="flex justify-between items-center gap-2">
            <p className="text-sm font-semibold truncate">{movie.title}</p>
            {watched && <Check className="size-4 text-sidebar-primary shrink-0" />}
          </div>
          <div className="flex justify-between items-center">
            <span className="text-xs text-muted-foreground">{movie.year}</span>
            {movie.rating > 0 && (
              <span className="text-xs font-semibold text-destructive">
                ⭐ {movie.rating.toFixed(1)}
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export function MovieCardSkeleton() {
  return (
    <div className="flex flex-col">
      <div className="bg-card border rounded-lg overflow-hidden flex flex-col">
        <div className="bg-muted w-full aspect-[2/3] animate-pulse" />
        <div className="p-3 flex flex-col gap-2">
          <div className="h-4 bg-muted rounded animate-pulse w-3/4" />
          <div className="flex justify-between">
            <div className="h-3 bg-muted rounded animate-pulse w-1/4" />
            <div className="h-3 bg-muted rounded animate-pulse w-1/4" />
          </div>
        </div>
      </div>
    </div>
  )
}

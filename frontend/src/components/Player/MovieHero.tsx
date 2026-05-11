import { FavoriteButton } from '@/components/page/FavoriteButton'
import { WatchLaterButton } from '@/components/page/WatchLaterButton'

interface MovieHeroProps {
  id: number
  title: string
  year: string
  runtime: string
  rating: number
  overview: string
  genres: string[]
  posterUrl: string
  backdropUrl: string
}

export function MovieHero({
  id,
  title,
  year,
  runtime,
  rating,
  overview,
  genres,
  posterUrl,
  backdropUrl,
}: MovieHeroProps) {
  return (
    <>
      {backdropUrl && (
        <div className="relative w-full h-56 sm:h-72 md:h-96 overflow-hidden">
          <img src={backdropUrl} alt="" className="w-full h-full object-cover" />
          <div className="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />
        </div>
      )}

      <div className={`flex gap-6 ${backdropUrl ? '-mt-24 relative z-10' : 'mt-8'}`}>
        {posterUrl && (
          <div className="hidden sm:block shrink-0 w-36 md:w-48">
            <img
              src={posterUrl}
              alt={title}
              className="w-full rounded-lg shadow-lg border border-border"
            />
          </div>
        )}

        <div className="flex flex-col gap-3 justify-end min-w-0">
          <h1 className="text-2xl md:text-4xl font-bold leading-tight">{title}</h1>

          <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
            {year && <span>{year}</span>}
            {runtime && <span>{runtime}</span>}
            {rating > 0 && (
              <span className="text-destructive font-semibold">⭐ {rating.toFixed(1)}</span>
            )}
          </div>

          {genres.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {genres.map(genre => (
                <span
                  key={genre}
                  className="text-xs px-2 py-0.5 rounded-full border border-border bg-muted text-muted-foreground"
                >
                  {genre}
                </span>
              ))}
            </div>
          )}

          {overview && (
            <p className="text-sm text-muted-foreground leading-relaxed max-w-2xl">{overview}</p>
          )}

          <div className="flex items-center gap-2 flex-wrap">
            <FavoriteButton tmdbId={id} />
            <WatchLaterButton tmdbId={id} />
          </div>
        </div>
      </div>
    </>
  )
}

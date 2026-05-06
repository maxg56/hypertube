'use client'

import { useState } from 'react'
import { Check } from 'lucide-react'

interface Film {
  year: string
  title: string
  poster?: string
  thumbnail?: string
  imdbRating?: number
}

const MOCK_FILMS: Film[] = [
  { year: '2023', title: 'Film 1', poster: 'https://robohash.org/1.png?set=set1', imdbRating: 8.2 },
  { year: '2022', title: 'Film 2', poster: 'https://robohash.org/2.png?set=set1', imdbRating: 7.5 },
  { year: '2023', title: 'Film 3', poster: 'https://robohash.org/3.png?set=set1', imdbRating: 8.9 },
  { year: '2021', title: 'Film 4', poster: 'https://robohash.org/4.png?set=set1', imdbRating: 7.1 },
  { year: '2023', title: 'Film 5', poster: 'https://robohash.org/5.png?set=set1', imdbRating: 8.7 },
  { year: '2022', title: 'Film 6', poster: 'https://robohash.org/6.png?set=set1', imdbRating: 6.8 },
  { year: '2023', title: 'Film 7', poster: 'https://robohash.org/7.png?set=set1', imdbRating: 8.4 },
  { year: '2021', title: 'Film 8', poster: 'https://robohash.org/8.png?set=set1', imdbRating: 7.9 },
  { year: '2023', title: 'Film 9', poster: 'https://robohash.org/9.png?set=set1', imdbRating: 8.1 },
  { year: '2022', title: 'Film 10', poster: 'https://robohash.org/10.png?set=set1', imdbRating: 7.3 },
  { year: '2023', title: 'Film 11', poster: 'https://robohash.org/11.png?set=set1', imdbRating: 8.6 },
  { year: '2021', title: 'Film 12', poster: 'https://robohash.org/12.png?set=set1', imdbRating: 7.7 },
  { year: '2023', title: 'Film 13', poster: 'https://robohash.org/13.png?set=set1', imdbRating: 8.3 },
  { year: '2022', title: 'Film 14', poster: 'https://robohash.org/14.png?set=set1', imdbRating: 7.2 },
  { year: '2023', title: 'Film 15', poster: 'https://robohash.org/15.png?set=set1', imdbRating: 8.8 },
  { year: '2021', title: 'Film 16', poster: 'https://robohash.org/16.png?set=set1', imdbRating: 7.4 },
  { year: '2023', title: 'Film 17', poster: 'https://robohash.org/17.png?set=set1', imdbRating: 8.5 },
  { year: '2022', title: 'Film 18', poster: 'https://robohash.org/18.png?set=set1', imdbRating: 7.6 },
  { year: '2023', title: 'Film 19', poster: 'https://robohash.org/19.png?set=set1', imdbRating: 8.0 },
  { year: '2021', title: 'Film 20', poster: 'https://robohash.org/20.png?set=set1', imdbRating: 7.8 },
  { year: '2023', title: 'Film 21', poster: 'https://robohash.org/21.png?set=set1', imdbRating: 8.4 },
  { year: '2022', title: 'Film 22', poster: 'https://robohash.org/22.png?set=set1', imdbRating: 7.1 },
  { year: '2023', title: 'Film 23', poster: 'https://robohash.org/23.png?set=set1', imdbRating: 8.9 },
  { year: '2021', title: 'Film 24', poster: 'https://robohash.org/24.png?set=set1', imdbRating: 7.5 },
  { year: '2023', title: 'Film 25', poster: 'https://robohash.org/25.png?set=set1', imdbRating: 8.2 },
  { year: '2022', title: 'Film 26', poster: 'https://robohash.org/26.png?set=set1', imdbRating: 6.9 },
  { year: '2023', title: 'Film 27', poster: 'https://robohash.org/27.png?set=set1', imdbRating: 8.7 },
  { year: '2021', title: 'Film 28', poster: 'https://robohash.org/28.png?set=set1', imdbRating: 7.9 },
  { year: '2023', title: 'Film 29', poster: 'https://robohash.org/29.png?set=set1', imdbRating: 8.3 },
  { year: '2022', title: 'Film 30', poster: 'https://robohash.org/30.png?set=set1', imdbRating: 7.4 },
  { year: '2023', title: 'Film 31', poster: 'https://robohash.org/31.png?set=set1', imdbRating: 8.6 },
  { year: '2021', title: 'Film 32', poster: 'https://robohash.org/32.png?set=set1', imdbRating: 7.7 },
  { year: '2023', title: 'Film 33', poster: 'https://robohash.org/33.png?set=set1', imdbRating: 8.1 },
  { year: '2022', title: 'Film 34', poster: 'https://robohash.org/34.png?set=set1', imdbRating: 7.2 },
  { year: '2023', title: 'Film 35', poster: 'https://robohash.org/35.png?set=set1', imdbRating: 8.5 },
  { year: '2021', title: 'Film 36', poster: 'https://robohash.org/36.png?set=set1', imdbRating: 7.3 },
]

export default function Thumbnails() {
  const [films] = useState<Film[]>(MOCK_FILMS)
  const [readFilms, setReadFilms] = useState<Set<number>>(new Set())

  const toggleRead = (index: number) => {
    const next = new Set(readFilms)
    if (next.has(index)) next.delete(index)
    else next.add(index)
    setReadFilms(next)
  }

  return (
    <div className="p-6">
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
        {films.map((film, index) => {
          const isRead = readFilms.has(index)
          return (
            <div
              key={index}
              className="group cursor-pointer flex flex-col"
              onClick={() => toggleRead(index)}
            >
              <div className={`card-glow bg-card border rounded-lg overflow-hidden flex flex-col flex-1 ${isRead ? 'opacity-50' : ''}`}>
                <div className="bg-muted w-full aspect-[2/3] overflow-hidden flex items-center justify-center">
                  {film.poster || film.thumbnail ? (
                    <img
                      src={film.poster ?? film.thumbnail}
                      alt={film.title}
                      className={`w-full h-full object-contain group-hover:scale-105 transition-transform ${isRead ? 'grayscale' : ''}`}
                    />
                  ) : (
                    <span className="text-muted-foreground text-center text-sm px-2">{film.title}</span>
                  )}
                </div>
                <div className="p-3 flex flex-col gap-1 flex-shrink-0">
                  <div className="flex justify-between items-center gap-2">
                    <p className="text-sm font-semibold truncate">{film.title}</p>
                    {isRead && <Check className="size-4 text-sidebar-primary shrink-0" />}
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-xs text-muted-foreground">{film.year}</span>
                    {film.imdbRating && (
                      <span className="text-xs font-semibold text-destructive">
                        ⭐ {film.imdbRating.toFixed(1)}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

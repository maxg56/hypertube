'use client'

import { useState } from 'react'

interface Film {
  id: string
  year: string
  title: string
  poster?: string
  thumbnail?: string
  imdbRating?: number
}

const MOCK_FILMS: Film[] = [
  { id: '1', year: '2023', title: 'Film 1', poster: 'https://robohash.org/1.png?set=set1', imdbRating: 8.2 },
  { id: '2', year: '2022', title: 'Film 2', poster: 'https://robohash.org/2.png?set=set1', imdbRating: 7.5 },
  { id: '3', year: '2023', title: 'Film 3', poster: 'https://robohash.org/3.png?set=set1', imdbRating: 8.9 },
  { id: '4', year: '2021', title: 'Film 4', poster: 'https://robohash.org/4.png?set=set1', imdbRating: 7.1 },
  { id: '5', year: '2023', title: 'Film 5', poster: 'https://robohash.org/5.png?set=set1', imdbRating: 8.7 },
  { id: '6', year: '2022', title: 'Film 6', poster: 'https://robohash.org/6.png?set=set1', imdbRating: 6.8 },
  { id: '7', year: '2023', title: 'Film 7', poster: 'https://robohash.org/7.png?set=set1', imdbRating: 8.4 },
  { id: '8', year: '2021', title: 'Film 8', poster: 'https://robohash.org/8.png?set=set1', imdbRating: 7.9 },
  { id: '9', year: '2023', title: 'Film 9', poster: 'https://robohash.org/9.png?set=set1', imdbRating: 8.1 },
  { id: '10', year: '2022', title: 'Film 10', poster: 'https://robohash.org/10.png?set=set1', imdbRating: 7.3 },
  { id: '11', year: '2023', title: 'Film 11', poster: 'https://robohash.org/11.png?set=set1', imdbRating: 8.6 },
  { id: '12', year: '2021', title: 'Film 12', poster: 'https://robohash.org/12.png?set=set1', imdbRating: 7.7 },
  { id: '13', year: '2023', title: 'Film 13', poster: 'https://robohash.org/13.png?set=set1', imdbRating: 8.3 },
  { id: '14', year: '2022', title: 'Film 14', poster: 'https://robohash.org/14.png?set=set1', imdbRating: 7.2 },
  { id: '15', year: '2023', title: 'Film 15', poster: 'https://robohash.org/15.png?set=set1', imdbRating: 8.8 },
  { id: '16', year: '2021', title: 'Film 16', poster: 'https://robohash.org/16.png?set=set1', imdbRating: 7.4 },
  { id: '17', year: '2023', title: 'Film 17', poster: 'https://robohash.org/17.png?set=set1', imdbRating: 8.5 },
  { id: '18', year: '2022', title: 'Film 18', poster: 'https://robohash.org/18.png?set=set1', imdbRating: 7.6 },
  { id: '19', year: '2023', title: 'Film 19', poster: 'https://robohash.org/19.png?set=set1', imdbRating: 8.0 },
  { id: '20', year: '2021', title: 'Film 20', poster: 'https://robohash.org/20.png?set=set1', imdbRating: 7.8 },
  { id: '21', year: '2023', title: 'Film 21', poster: 'https://robohash.org/21.png?set=set1', imdbRating: 8.4 },
  { id: '22', year: '2022', title: 'Film 22', poster: 'https://robohash.org/22.png?set=set1', imdbRating: 7.1 },
  { id: '23', year: '2023', title: 'Film 23', poster: 'https://robohash.org/23.png?set=set1', imdbRating: 8.9 },
  { id: '24', year: '2021', title: 'Film 24', poster: 'https://robohash.org/24.png?set=set1', imdbRating: 7.5 },
  { id: '25', year: '2023', title: 'Film 25', poster: 'https://robohash.org/25.png?set=set1', imdbRating: 8.2 },
  { id: '26', year: '2022', title: 'Film 26', poster: 'https://robohash.org/26.png?set=set1', imdbRating: 6.9 },
  { id: '27', year: '2023', title: 'Film 27', poster: 'https://robohash.org/27.png?set=set1', imdbRating: 8.7 },
  { id: '28', year: '2021', title: 'Film 28', poster: 'https://robohash.org/28.png?set=set1', imdbRating: 7.9 },
  { id: '29', year: '2023', title: 'Film 29', poster: 'https://robohash.org/29.png?set=set1', imdbRating: 8.3 },
  { id: '30', year: '2022', title: 'Film 30', poster: 'https://robohash.org/30.png?set=set1', imdbRating: 7.4 },
  { id: '31', year: '2023', title: 'Film 31', poster: 'https://robohash.org/31.png?set=set1', imdbRating: 8.6 },
  { id: '32', year: '2021', title: 'Film 32', poster: 'https://robohash.org/32.png?set=set1', imdbRating: 7.7 },
  { id: '33', year: '2023', title: 'Film 33', poster: 'https://robohash.org/33.png?set=set1', imdbRating: 8.1 },
  { id: '34', year: '2022', title: 'Film 34', poster: 'https://robohash.org/34.png?set=set1', imdbRating: 7.2 },
  { id: '35', year: '2023', title: 'Film 35', poster: 'https://robohash.org/35.png?set=set1', imdbRating: 8.5 },
  { id: '36', year: '2021', title: 'Film 36', poster: 'https://robohash.org/36.png?set=set1', imdbRating: 7.3 },
]

export default function Thumbnails() {
  const [films] = useState<Film[]>(MOCK_FILMS)
  const [readFilms, setReadFilms] = useState<Set<number>>(new Set())

  const toggleRead = (index: number) => {
    const newReadFilms = new Set(readFilms)
    if (newReadFilms.has(index)) {
      newReadFilms.delete(index)
    } else {
      newReadFilms.add(index)
    }
    setReadFilms(newReadFilms)
  }

  return (
    <div className="w-full bg-gradient-to-t from-orange-400 to-blue-400 p-6 shadow-lg min-h-screen">
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 w-full">
        {films.map((film, index) => {
          const isRead = readFilms.has(index)
          return (
          <div key={index} className="group cursor-pointer flex flex-col h-full" onClick={() => toggleRead(index)}>
            <div className={`bg-gray-800 rounded-lg overflow-hidden hover:shadow-blue transition-all flex flex-col flex-1 ${isRead ? 'opacity-50' : ''}`}>
              <div className="bg-gray-300 w-full flex-1 min-h-0 overflow-hidden flex items-center justify-center">
                {film.poster || film.thumbnail ? (
                  <img
                    src={film.poster || film.thumbnail}
                    alt={film.title}
                    className={`w-full h-full object-contain group-hover:scale-105 transition-transform ${isRead ? 'grayscale' : ''}`}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center bg-gray-800">
                    <span className="text-gray-500 text-center px-2">{film.title}</span>
                  </div>
                )}
              </div>

              <div className="p-3 bg-gray-900 text-white flex flex-col gap-2 flex-shrink-0">
                <div className="flex justify-between items-center">
                  <p className="text-sm font-bold truncate">{film.title}</p>
                  {isRead && <div className="text-white text-lg">✓</div>}
                </div>
                <div className="flex justify-between items-center">
                  <p className="text-xs text-gray-300"><span className="font-semibold">{film.year}</span></p>
                  {film.imdbRating && (
                    <div className="text-sm font-semibold text-yellow-400 flex items-center">
                      <span className="mr-1">⭐</span>
                      {film.imdbRating.toFixed(1)}
                    </div>
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
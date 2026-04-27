'use client'

import { useState } from 'react'

interface Film {
  id: string
  title: string
  poster?: string
  thumbnail?: string
}

const MOCK_FILMS: Film[] = [
  { id: '1', title: 'Film 1', poster: 'https://robohash.org/1.png?set=set1' },
  { id: '2', title: 'Film 2', poster: 'https://robohash.org/2.png?set=set1' },
  { id: '3', title: 'Film 3', poster: 'https://robohash.org/3.png?set=set1' },
  { id: '4', title: 'Film 4', poster: 'https://robohash.org/4.png?set=set1' },
  { id: '5', title: 'Film 5', poster: 'https://robohash.org/5.png?set=set1' },
  { id: '6', title: 'Film 6', poster: 'https://robohash.org/6.png?set=set1' },
  { id: '7', title: 'Film 7', poster: 'https://robohash.org/7.png?set=set1' },
  { id: '8', title: 'Film 8', poster: 'https://robohash.org/8.png?set=set1' },
  { id: '9', title: 'Film 9', poster: 'https://robohash.org/9.png?set=set1' },
  { id: '10', title: 'Film 10', poster: 'https://robohash.org/10.png?set=set1' },
  { id: '11', title: 'Film 11', poster: 'https://robohash.org/11.png?set=set1' },
  { id: '12', title: 'Film 12', poster: 'https://robohash.org/12.png?set=set1' },
  { id: '13', title: 'Film 13', poster: 'https://robohash.org/13.png?set=set1' },
  { id: '14', title: 'Film 14', poster: 'https://robohash.org/14.png?set=set1' },
  { id: '15', title: 'Film 15', poster: 'https://robohash.org/15.png?set=set1' },
  { id: '16', title: 'Film 16', poster: 'https://robohash.org/16.png?set=set1' },
  { id: '17', title: 'Film 17', poster: 'https://robohash.org/17.png?set=set1' },
  { id: '18', title: 'Film 18', poster: 'https://robohash.org/18.png?set=set1' },
  { id: '19', title: 'Film 19', poster: 'https://robohash.org/19.png?set=set1' },
  { id: '20', title: 'Film 20', poster: 'https://robohash.org/20.png?set=set1' },
  { id: '21', title: 'Film 21', poster: 'https://robohash.org/21.png?set=set1' },
  { id: '22', title: 'Film 22', poster: 'https://robohash.org/22.png?set=set1' },
  { id: '23', title: 'Film 23', poster: 'https://robohash.org/23.png?set=set1' },
  { id: '24', title: 'Film 24', poster: 'https://robohash.org/24.png?set=set1' },
  { id: '25', title: 'Film 25', poster: 'https://robohash.org/25.png?set=set1' },
  { id: '26', title: 'Film 26', poster: 'https://robohash.org/26.png?set=set1' },
  { id: '27', title: 'Film 27', poster: 'https://robohash.org/27.png?set=set1' },
  { id: '28', title: 'Film 28', poster: 'https://robohash.org/28.png?set=set1' },
  { id: '29', title: 'Film 29', poster: 'https://robohash.org/29.png?set=set1' },
  { id: '30', title: 'Film 30', poster: 'https://robohash.org/30.png?set=set1' },
  { id: '31', title: 'Film 31', poster: 'https://robohash.org/31.png?set=set1' },
  { id: '32', title: 'Film 32', poster: 'https://robohash.org/32.png?set=set1' },
  { id: '33', title: 'Film 33', poster: 'https://robohash.org/33.png?set=set1' },
  { id: '34', title: 'Film 34', poster: 'https://robohash.org/34.png?set=set1' },
  { id: '35', title: 'Film 35', poster: 'https://robohash.org/35.png?set=set1' },
  { id: '36', title: 'Film 36', poster: 'https://robohash.org/36.png?set=set1' },
]

export default function Thumbnails() {
  const [films] = useState<Film[]>(MOCK_FILMS)

  return (
    <div className="w-1040  bg-gradient-to-t from-orange-400 to-blue-400 p-6 shadow-lg">
      <div className="grid grid-cols-5 gap-4">
        {films.map((film) => (
          <div key={film.id} className="group cursor-pointer">
            <div className="bg-gray-300 h-64 rounded-lg overflow-hidden hover:shadow-lg transition-shadow">
              {film.poster || film.thumbnail ? (
                <img
                  src={film.poster || film.thumbnail}
                  alt={film.title}
                  className="w-full h-full object-cover group-hover:scale-105 transition-transform"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center bg-gray-200">
                  <span className="text-gray-500 text-center px-2">{film.title}</span>
                </div>
              )}
            </div>
            <p className="mt-2 text-sm font-medium truncate">{film.title}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
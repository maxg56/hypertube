# library-service

Fournit les métadonnées films et les liens torrent pour Hypertube. Agrège trois sources : TMDb (métadonnées riches), OMDb (fallback), et YTS (torrents).

## Responsibilities

- Recherche de films via TMDb ou OMDb (fallback)
- Détail d'un film enrichi avec les torrents YTS
- Recherche directe sur YTS (films disponibles en torrent avec seeds/peers)
- Cache Redis 24h pour les métadonnées, 1h pour les résultats YTS

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | — | Health check |
| `GET` | `/api/v1/library/movies/search` | JWT | Recherche TMDb / OMDb |
| `GET` | `/api/v1/library/movies/yts` | JWT | Recherche YTS (avec torrents) |
| `GET` | `/api/v1/library/movies/:id` | JWT | Détail d'un film (TMDb + torrents YTS) |

---

### GET /api/v1/library/movies/search

| Paramètre | Type | Requis | Description |
|-----------|------|--------|-------------|
| `q` | string | oui | Terme de recherche |
| `page` | int | non | Page (défaut : 1) |

```json
// 200 OK
{
  "success": true,
  "data": {
    "page": 1,
    "total_pages": 5,
    "results": [
      {
        "id": 27205,
        "title": "Inception",
        "year": "2010",
        "overview": "Cobb, a skilled thief...",
        "rating": 8.4,
        "poster_url": "https://image.tmdb.org/t/p/w500/...",
        "backdrop_url": "https://image.tmdb.org/t/p/w500/...",
        "source": "tmdb"
      }
    ]
  }
}
```

Erreurs : `400` paramètre `q` manquant · `502` provider injoignable · `503` aucun provider configuré

---

### GET /api/v1/library/movies/yts

| Paramètre | Type | Requis | Description |
|-----------|------|--------|-------------|
| `q` | string | oui | Terme de recherche |
| `page` | int | non | Page (défaut : 1) |

Retourne les films disponibles sur YTS triés par nombre de seeds, avec les liens torrent inclus.

```json
// 200 OK
{
  "success": true,
  "data": {
    "page": 1,
    "total_pages": 3,
    "results": [
      {
        "id": 15442,
        "imdb_id": "tt1375666",
        "title": "Inception",
        "year": "2010",
        "runtime": 148,
        "rating": 8.8,
        "poster_url": "https://yts.mx/assets/images/movies/...",
        "genres": ["Action", "Adventure", "Sci-Fi"],
        "source": "yts",
        "torrents": [
          {
            "url": "https://yts.mx/torrent/download/...",
            "hash": "ABC123...",
            "quality": "1080p",
            "type": "bluray",
            "size": "2.18 GB",
            "seeds": 4210,
            "peers": 312
          }
        ]
      }
    ]
  }
}
```

Erreurs : `400` paramètre `q` manquant · `502` YTS injoignable

---

### GET /api/v1/library/movies/:id

`:id` est l'ID numérique TMDb du film (ex : `27205` pour Inception).

Récupère les métadonnées complètes depuis TMDb (titre, résumé, casting, genres, durée, note, images) puis enrichit automatiquement la réponse avec les torrents YTS via l'IMDb ID.

```json
// 200 OK
{
  "success": true,
  "data": {
    "id": 27205,
    "imdb_id": "tt1375666",
    "title": "Inception",
    "year": "2010",
    "overview": "Cobb, a skilled thief who commits corporate espionage...",
    "runtime": 148,
    "rating": 8.4,
    "poster_url": "https://image.tmdb.org/t/p/w500/...",
    "backdrop_url": "https://image.tmdb.org/t/p/w500/...",
    "genres": ["Action", "Adventure", "Science Fiction", "Mystery"],
    "cast": [
      { "name": "Leonardo DiCaprio", "character": "Cobb", "order": 0 },
      { "name": "Joseph Gordon-Levitt", "character": "Arthur", "order": 1 }
    ],
    "source": "tmdb",
    "torrents": [
      {
        "url": "https://yts.mx/torrent/download/...",
        "hash": "ABC123...",
        "quality": "1080p",
        "type": "bluray",
        "size": "2.18 GB",
        "seeds": 4210,
        "peers": 312
      }
    ]
  }
}
```

Erreurs : `400` ID invalide · `404` film introuvable · `502` TMDb injoignable

---

## Sources de données

| Source | Clé API | Rôle |
|--------|---------|------|
| TMDb | `TMDB_API_KEY` (requis) | Métadonnées principales (recherche, détail, casting) |
| OMDb | `OMDB_API_KEY` (optionnel) | Fallback si TMDb indisponible |
| YTS | — (public) | Torrents (hash, qualité, seeds, peers) |

La logique de fallback pour la recherche : TMDb → OMDb (si TMDb KO). Pour le détail d'un film, TMDb est utilisé pour les métadonnées, YTS pour les torrents (via IMDb ID).

---

## Cache Redis

| Clé | TTL | Contenu |
|-----|-----|---------|
| `search:{q}:page:{n}` | 24h | Résultats de recherche TMDb/OMDb |
| `movie:{tmdb_id}` | 24h | Détail film enrichi (métadonnées + torrents) |
| `yts:search:{q}:page:{n}` | 1h | Résultats YTS (TTL court — seeds varient) |

Redis est optionnel : si indisponible au démarrage, le service fonctionne sans cache.

---

## Environment variables

| Variable | Défaut | Description |
|----------|--------|-------------|
| `PORT` | `8003` | Port d'écoute |
| `TMDB_API_KEY` | — | **Requis.** Clé API TMDb |
| `OMDB_API_KEY` | — | Clé API OMDb (fallback optionnel) |
| `REDIS_HOST` | `localhost` | Host Redis |
| `REDIS_PORT` | `6379` | Port Redis |
| `REDIS_PASSWORD` | — | Mot de passe Redis |

---

## Layout

```
src/
├── client/
│   ├── tmdb.go     # Client HTTP TMDb (search + detail + credits)
│   ├── omdb.go     # Client HTTP OMDb (search + title lookup)
│   └── yts.go      # Client HTTP YTS (search + lookup par IMDb ID)
├── conf/
│   └── redis.go    # Connexion Redis + helpers GetCache/SetCache
├── handlers/
│   └── movie.go    # Handlers Search, GetMovie, SearchYTS
├── models/
│   └── movie.go    # Movie, CastMember, Torrent, SearchResult
├── utils/
│   └── response.go # RespondSuccess / RespondError
└── main.go         # Routes Gin + init Redis
```

## Development

```bash
docker compose -f docker-compose.dev.yml up library-service
```

Source files sont volume-montés — Air recharge à chaque modification `.go`.

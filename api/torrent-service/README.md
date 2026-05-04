# torrent-service

Gère le téléchargement BitTorrent à la demande et le streaming vidéo HTTP pour Hypertube.

## Responsabilités

- Démarrage d'un téléchargement torrent via magnet URI (premier utilisateur qui demande le film)
- Téléchargement non-bloquant — le streaming commence dès les premiers morceaux reçus
- Suivi de la progression en base de données (toutes les 5 s)
- Reprise automatique des téléchargements en cours au redémarrage du service
- Streaming HTTP avec support des requêtes `Range` (seek, lecture simultanée multi-clients)
- Sauvegarde persistante du fichier une fois le téléchargement terminé

## Endpoints

| Méthode | Chemin | Auth | Description |
|---------|--------|------|-------------|
| `GET`  | `/health` | — | Health check |
| `POST` | `/api/v1/torrent/download` | JWT | Démarre (ou retrouve) un téléchargement |
| `GET`  | `/api/v1/torrent/status/:id` | JWT | Progression d'un téléchargement |
| `GET`  | `/api/v1/stream/:id` | JWT | Stream du fichier vidéo (Range supporté) |

---

### POST /api/v1/torrent/download

Démarre le téléchargement d'un torrent. Idempotent : si le hash est déjà connu, renvoie l'état existant sans relancer.

```json
// Requête
{
  "magnet_uri": "magnet:?xt=urn:btih:<hash>&dn=<name>&tr=...",
  "movie_id":   42
}

// 202 Accepted
{
  "success": true,
  "data": {
    "info_hash": "da39a3ee5e6b4b0d3255bfef95601890afd80709",
    "status":    "downloading",
    "message":   "torrent download started"
  }
}
```

Erreurs : `400` magnet URI invalide · `400` champs manquants · `400` client non initialisé

---

### GET /api/v1/torrent/status/:id

Retourne la progression d'un téléchargement. `:id` est le hash hexadécimal en minuscules.

```json
// 200 OK
{
  "success": true,
  "data": {
    "info_hash":        "da39a3ee5e6b4b0d3255bfef95601890afd80709",
    "status":           "downloading",
    "progress":         34.72,
    "downloaded_bytes": 364904448,
    "file_size_bytes":  1073741824,
    "file_path":        ""
  }
}
```

| Champ | Description |
|-------|-------------|
| `status` | `pending` · `downloading` · `ready` · `error` |
| `progress` | Pourcentage (0–100) |
| `file_path` | Chemin absolu du fichier, renseigné quand `status = ready` |
| `error` | Message d'erreur, renseigné quand `status = error` |

Erreurs : `404` hash inconnu

---

### GET /api/v1/stream/:id

Streame le fichier vidéo. `:id` est le hash hexadécimal du torrent.

- Supporte l'en-tête `Range` → réponse `206 Partial Content` (seek et reprise)
- Plusieurs clients peuvent streamer le même torrent simultanément
- Pendant le téléchargement, les lectures bloquent naturellement sur les morceaux non encore reçus

```
GET /api/v1/stream/da39a3ee5e6b4b0d3255bfef95601890afd80709
Range: bytes=0-

HTTP/1.1 206 Partial Content
Content-Type: video/x-matroska
Accept-Ranges: bytes
Content-Range: bytes 0-1073741823/1073741824
```

Codes de retour :

| Code | Cas |
|------|-----|
| `206` | Succès, contenu partiel (Range) |
| `200` | Succès, contenu complet |
| `202` | Torrent `pending`, réessayer après `Retry-After: 5` s |
| `404` | Hash inconnu |
| `503` | Torrent en erreur ou fichier inaccessible |

---

## Architecture interne

```
POST /download
    └─ StartDownload()
          ├─ extractInfoHash()          parse le magnet, normalise en lowercase
          ├─ findOrCreateRecord()       upsert DB (idempotent)
          ├─ client.AddMagnet()         ajoute au client anacrolix/torrent
          └─ go monitorTorrent()        goroutine de suivi en arrière-plan

monitorTorrent()
    ├─ GotInfo() ─── timeout 2 min → status=error
    ├─ DownloadAll() + prioritizeForStreaming()
    │       premiers 5 Mio → PiecePriorityNow (lecture immédiate)
    │       reste → PiecePriorityNormal
    ├─ ticker 5 s → UPDATE downloaded, progress
    ├─ watchdog 10 min sans progrès → status=error
    └─ complet → status=ready, file_path renseigné

GET /stream/:id
    └─ GetTorrentReader()
          ├─ torrent actif  → file.NewReader() (bloque sur morceaux manquants)
          └─ torrent terminé → os.Open(file_path)
                └─ http.ServeContent() — gère Range/206 automatiquement
```

Le client `anacrolix/torrent` est un singleton partagé entre tous les téléchargements. Au démarrage, `reattachPendingTorrents()` recharge automatiquement les torrents dont le statut est `pending` ou `downloading`.

## Structure des fichiers

```
src/
├── main.go
├── conf/
│   └── db.go                 connexion PostgreSQL
├── models/
│   └── torrent.go            modèle GORM → table torrents
├── types/
│   └── request_types.go      structs requête / réponse
├── services/
│   ├── client.go             singleton anacrolix, init, reattach
│   ├── download.go           StartDownload, extractInfoHash, addToClient
│   ├── monitor.go            monitorTorrent, prioritizeForStreaming, setError
│   └── reader.go             GetTorrentReader, GetRecord
├── handlers/
│   ├── download.go           POST /api/v1/torrent/download
│   ├── status.go             GET  /api/v1/torrent/status/:id
│   └── stream.go             GET  /api/v1/stream/:id
└── utils/
    └── response.go           enveloppe JSON standard
```

## Variables d'environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `PORT` | `8004` | Port d'écoute |
| `TORRENT_DOWNLOAD_PATH` | `/data/torrents` | Répertoire de stockage des fichiers |
| `STREAM_BUFFER_SIZE` | — | Taille du buffer de stream (MiB, non utilisé directement) |
| `DB_HOST` | `localhost` | Hôte PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_NAME` | `hypertube` | Nom de la base |
| `DB_USER` | `postgres` | Utilisateur |
| `DB_PASSWORD` | `password` | Mot de passe |

## Tests

```bash
go test ./src/services/... ./src/handlers/... -v
```

Les tests utilisent SQLite en mémoire (`github.com/glebarez/sqlite`, pur Go) pour éviter le conflit de symboles C avec `anacrolix/torrent`.

| Package | Tests |
|---------|-------|
| `services` | `extractInfoHash`, `findOrCreateRecord`, `GetRecord`, `StartDownload` |
| `handlers` | health check, download (validation + erreurs), status (tous statuts), stream (pending/error/fichier manquant) |

## Gestion des erreurs

| Cas | Comportement |
|-----|-------------|
| Magnet URI invalide | `400` immédiat |
| Pas de seeders / peers | `GotInfo()` expire après 2 min → `status=error` |
| Téléchargement bloqué | Watchdog 10 min sans progrès → `status=error` |
| Requête dupliquée | Idempotent — renvoie le hash et le statut courant |
| Redémarrage du service | `reattachPendingTorrents()` reprend les téléchargements en cours |

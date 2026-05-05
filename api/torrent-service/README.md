# torrent-service

Gère le téléchargement BitTorrent à la demande et le streaming vidéo HTTP pour Hypertube.

## Responsabilités

- Démarrage d'un téléchargement torrent via magnet URI (premier utilisateur qui demande le film)
- Téléchargement non-bloquant — le streaming commence dès les premiers morceaux reçus
- Suivi de la progression en base de données (toutes les 5 s)
- Reprise automatique des téléchargements en cours au redémarrage du service
- Streaming HTTP avec support des requêtes `Range` pour les formats natifs (seek, lecture simultanée multi-clients)
- Transcodage à la volée des formats non supportés nativement (MKV, AVI, MOV…) vers MP4 via ffmpeg
- Sauvegarde persistante du fichier une fois le téléchargement terminé

## Endpoints

| Méthode | Chemin | Auth | Description |
|---------|--------|------|-------------|
| `GET`  | `/health` | — | Health check |
| `POST` | `/api/v1/torrent/download` | JWT | Démarre (ou retrouve) un téléchargement |
| `GET`  | `/api/v1/torrent/status/:id` | JWT | Progression d'un téléchargement |
| `GET`  | `/api/v1/stream/:id` | JWT | Stream du fichier vidéo (transcodage automatique) |

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

Le service détecte automatiquement si le format nécessite un transcodage :

**Format natif (MP4, WebM, OGG, M4V)**
- Servi directement avec `http.ServeContent`
- Supporte l'en-tête `Range` → réponse `206 Partial Content` (seek, reprise)

```
GET /api/v1/stream/da39a3ee5e6b4b0d3255bfef95601890afd80709
Range: bytes=0-

HTTP/1.1 206 Partial Content
Content-Type: video/mp4
Accept-Ranges: bytes
Content-Range: bytes 0-1073741823/1073741824
```

**Format non natif (MKV, AVI, MOV…)**
- Transcodage à la volée via ffmpeg (pipeline torrent reader → ffmpeg stdin → HTTP response, sans fichier temporaire)
- Si les codecs source sont déjà H.264 + AAC/MP3 : remux uniquement (`-c copy`), pas de ré-encodage
- Sinon : ré-encodage complet en H.264/AAC (`-preset ultrafast -tune zerolatency`)
- Sortie en **fragmented MP4** (`frag_keyframe+empty_moov`), compatible Firefox et Chrome
- `Range` non supporté — le navigateur reçoit un stream continu (`Transfer-Encoding: chunked`)

```
GET /api/v1/stream/da39a3ee5e6b4b0d3255bfef95601890afd80709

HTTP/1.1 200 OK
Content-Type: video/mp4
Cache-Control: no-cache
Transfer-Encoding: chunked
```

Codes de retour :

| Code | Cas |
|------|-----|
| `206` | Format natif avec `Range` |
| `200` | Format natif sans `Range`, ou format transcodé |
| `202` | Torrent `pending`, réessayer après `Retry-After: 5` s |
| `404` | Hash inconnu |
| `500` | Erreur de lancement du transcodage ffmpeg |
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
    └─ NeedsTranscoding(filename)
          ├─ format natif (mp4/webm/ogg/m4v)
          │     └─ http.ServeContent() — gère Range/206 automatiquement
          └─ format non natif (mkv/avi/mov…)
                ├─ ProbeCodecs() via ffprobe (si fichier sur disque)
                ├─ canCopyStream() → remux -c copy si H.264+AAC, sinon libx264+aac
                └─ StartTranscode() → ffmpeg pipe:0→pipe:1 → HTTP chunked
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
│   ├── reader.go             GetTorrentReader, GetRecord
│   └── transcoder.go         NeedsTranscoding, ProbeCodecs, StartTranscode
├── handlers/
│   ├── download.go           POST /api/v1/torrent/download
│   ├── status.go             GET  /api/v1/torrent/status/:id
│   └── stream.go             GET  /api/v1/stream/:id
└── utils/
    └── response.go           enveloppe JSON standard
```

## Dépendances système

| Outil | Rôle |
|-------|------|
| `ffmpeg` | Transcodage à la volée des formats non natifs |
| `ffprobe` | Détection des codecs source (remux vs ré-encodage) |

Les deux binaires doivent être présents dans le `PATH` du conteneur. Si `ffmpeg` est absent, les requêtes de stream sur des formats non natifs renvoient `500`.

## Variables d'environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `PORT` | `8004` | Port d'écoute |
| `TORRENT_DOWNLOAD_PATH` | `/data/torrents` | Répertoire de stockage des fichiers |
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
| `services` | `extractInfoHash`, `findOrCreateRecord`, `GetRecord`, `StartDownload`, `NeedsTranscoding`, `canCopyStream`, `buildFFmpegArgs`, `ProbeCodecs`, `StartTranscode` (intégration, requiert ffmpeg/ffprobe) |
| `handlers` | health check, download (validation + erreurs), status (tous statuts), stream (pending/error/fichier manquant/succès/Range 206/MIME types/transcodage MKV→MP4) |

## Gestion des erreurs

| Cas | Comportement |
|-----|-------------|
| Magnet URI invalide | `400` immédiat |
| Pas de seeders / peers | `GotInfo()` expire après 2 min → `status=error` |
| Téléchargement bloqué | Watchdog 10 min sans progrès → `status=error` |
| Requête dupliquée | Idempotent — renvoie le hash et le statut courant |
| Redémarrage du service | `reattachPendingTorrents()` reprend les téléchargements en cours |

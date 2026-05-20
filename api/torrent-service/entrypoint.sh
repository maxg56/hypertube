#!/bin/sh
set -e

mkdir -p /data/torrents /data/subtitles
chmod 777 /data/torrents /data/subtitles

if [ "$(id -u)" = "0" ] && command -v su-exec >/dev/null 2>&1; then
    exec su-exec torrentservice "$@"
else
    exec "$@"
fi

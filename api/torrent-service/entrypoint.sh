#!/bin/sh
set -e

mkdir -p /data/torrents /data/subtitles
exec "$@"

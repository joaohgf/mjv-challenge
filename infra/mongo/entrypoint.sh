#!/bin/sh
set -eu

key_file=/data/configdb/keyfile

if [ ! -f "$key_file" ]; then
  umask 077
  head -c 756 /dev/urandom | base64 > "$key_file"
  chown mongodb:mongodb "$key_file"
fi

exec /usr/local/bin/docker-entrypoint.sh "$@"

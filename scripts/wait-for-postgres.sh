#!/usr/bin/env bash
host="$1"
port="$2"

until pg_isready -h "$host" -p "$port" > /dev/null 2>&1; do
  >&2 echo "Postgres is unavailable – sleeping"
  sleep 1
done

>&2 echo "Postgres is up – executing command"
exec "$@"

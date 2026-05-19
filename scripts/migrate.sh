#!/bin/sh
# Run database migrations against DATABASE_URL.
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "ERROR: DATABASE_URL is not set" >&2
  exit 1
fi

echo "Running migrations..."
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
echo "Migrations complete."

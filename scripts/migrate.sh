#!/bin/sh
# Run database migrations against DATABASE_URL.
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "ERROR: DATABASE_URL is not set" >&2
  exit 1
fi

echo "Running migrations..."
for f in migrations/*.sql; do
  echo "  applying $f"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done
echo "Migrations complete."

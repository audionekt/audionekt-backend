#!/bin/bash

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -p 5432 -U dev; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is ready!"

# Run migrations
echo "Running database migrations..."
psql -h localhost -p 5432 -U dev -d musicapp -f /docker-entrypoint-initdb.d/001_initial_schema.sql

echo "Database initialization complete!"

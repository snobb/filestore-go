#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <migration_name>"
    exit 1
fi

docker build -t filestore-migrations:latest -f Dockerfile.migrations .
docker run --rm --network=host -e DATABASE_URL -v "$(pwd)/db/migrations:/app/db/migrations" filestore-migrations new "$1"

#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <migration_name>"
    exit 1
fi

podman build -t filestore-migrations:latest -f Containerfile.migrations .
podman run --rm --network=host -e DATABASE_URL -v "$(pwd)/db/migrations:/app/db/migrations" filestore-migrations new "$1"

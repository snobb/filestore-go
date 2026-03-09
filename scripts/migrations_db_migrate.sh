#!/bin/bash
set -e

docker build -t filestore-migrations:latest -f Dockerfile.migrations .
docker run --rm --network=host -e DATABASE_URL filestore-migrations up

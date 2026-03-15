#!/bin/bash
set -e

podman build -t filestore-migrations:latest -f Containerfile.migrations .
podman run --rm --network=host -e DATABASE_URL filestore-migrations up

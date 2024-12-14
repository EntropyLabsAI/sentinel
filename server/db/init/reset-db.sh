#!/bin/bash

# Stop any running containers
docker compose down

# Remove the volume
docker volume rm -f asteroid_postgres_data || true

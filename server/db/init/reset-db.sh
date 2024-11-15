#!/bin/bash

# Stop all containers and remove volumes
docker compose down -v

# Remove the postgres volume explicitly
docker volume rm -f sentinel_postgres_data || true

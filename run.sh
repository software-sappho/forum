#!/bin/bash

# Ensure forum.db exists as a file (not a directory)
if [ -d "forum.db" ]; then
  echo "Removing forum.db directory..."
  rm -rf forum.db
fi

touch forum.db

echo "forum.db file ensured."

echo "Sourcing API keys..."
source apiKeys.sh

echo "Building and starting Docker containers..."
docker-compose up -d

echo "Checking container status..."
docker-compose ps

echo "Tailing logs (press Ctrl+C to exit)..."
docker-compose logs -f 
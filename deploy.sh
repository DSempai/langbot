#!/bin/bash

# Simple VPS deployment script
# 1. Copy env.example to .env and add your bot token
# 2. Run this script

# Create data directory for persistent storage
mkdir -p data

# Copy database if it doesn't exist in data directory
if [ ! -f data/dutch_learning.db ]; then
    cp dutch_learning.db data/
fi

# Build and start the bot
docker compose up -d --build

echo "Bot deployed! Check logs with: docker-compose logs -f"
echo "Stop with: docker-compose down" 
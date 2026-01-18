#!/bin/sh
set -e

cd /app

echo "==> Installing frontend dependencies..."
cd frontend
bun install
cd ..

echo "==> Starting development servers..."

# Start Vite dev server in background
cd frontend
bun run dev &
VITE_PID=$!
cd ..

# Give Vite a moment to start
sleep 2

echo "==> Starting Go backend with Air hot-reload..."

# Start Air for Go hot-reload
air -c .air.toml

# If Air exits, clean up Vite
kill $VITE_PID 2>/dev/null || true

#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"

USE_FLIGHT_MOCK=0
BUILD_MODE="run"

usage() {
  cat <<'EOF'
Usage:
  ./start_all.sh [--use-flight-mock] [--build]

Options:
  --use-flight-mock   Start mcp-flight mock on localhost:8082
  --build             Build binaries first, then start them
  -h, --help          Show this help message

Examples:
  ./start_all.sh
  ./start_all.sh --use-flight-mock
  ./start_all.sh --use-flight-mock --build

Ports:
  mcp-time    -> localhost:8081
  mcp-flight  -> localhost:8082
  mcp-chatbox -> localhost:8083
  mcp-photo   -> localhost:8084

Notes:
  - Real mcp-flight needs SERPAPI_API_KEY
  - Real mcp-photo needs UNSPLASH_ACCESS_KEY
  - mcp-time google_search needs GOOGLE_API_KEY and GOOGLE_SEARCH_ENGINE_ID
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --use-flight-mock)
      USE_FLIGHT_MOCK=1
      shift
      ;;
    --build)
      BUILD_MODE="build"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

mkdir -p "$LOG_DIR"

PIDS=()

cleanup() {
  if [[ ${#PIDS[@]} -gt 0 ]]; then
    echo
    echo "Stopping MCP services..."
    kill "${PIDS[@]}" 2>/dev/null || true
    wait "${PIDS[@]}" 2>/dev/null || true
  fi
}

trap cleanup INT TERM EXIT

start_service_build() {
  local name="$1"
  local workdir="$2"
  local binary_name="$3"
  local logfile="$4"
  shift 4

  echo "Building $name ..."
  (
    cd "$workdir"
    go build -o "$binary_name" main.go
  )

  echo "Starting $name ..."
  (
    cd "$workdir"
    exec "./$binary_name" "$@"
  ) >"$logfile" 2>&1 &
  PIDS+=("$!")
}

start_service() {
  local name="$1"
  local workdir="$2"
  local binary_name="$3"
  local logfile="$4"
  shift 4

  if [[ "$BUILD_MODE" == "build" ]]; then
    start_service_build "$name" "$workdir" "$binary_name" "$logfile" "$@"
  else
    (
      cd "$workdir"
      exec go run . "$@"
    ) >"$logfile" 2>&1 &
    PIDS+=("$!")
  fi
}

TIME_DIR="$SCRIPT_DIR/mcp-time"
FLIGHT_DIR="$SCRIPT_DIR/mcp-flight"
FLIGHT_MOCK_DIR="$SCRIPT_DIR/mcp-flight/mock"
CHATBOX_DIR="$SCRIPT_DIR/mcp-chatbox"
PHOTO_DIR="$SCRIPT_DIR/mcp-photo"

start_service \
  "mcp-time" \
  "$TIME_DIR" \
  "mcp-time" \
  "$LOG_DIR/mcp-time.log" \
  -transport=sse -server_listen=localhost:8081

if [[ "$USE_FLIGHT_MOCK" == "1" ]]; then
  start_service \
    "mcp-flight-mock" \
    "$FLIGHT_MOCK_DIR" \
    "mcp-flight-mock" \
    "$LOG_DIR/mcp-flight.log" \
    -transport=sse -server_listen=localhost:8082
else
  start_service \
    "mcp-flight" \
    "$FLIGHT_DIR" \
    "mcp-flight" \
    "$LOG_DIR/mcp-flight.log" \
    -transport=sse -server_listen=localhost:8082
fi

start_service \
  "mcp-chatbox" \
  "$CHATBOX_DIR" \
  "mcp-chatbox" \
  "$LOG_DIR/mcp-chatbox.log" \
  -transport=sse -server_listen=localhost:8083

start_service \
  "mcp-photo" \
  "$PHOTO_DIR" \
  "mcp-photo" \
  "$LOG_DIR/mcp-photo.log" \
  -transport=sse -server_listen=localhost:8084

echo
echo "MCP services started."
echo "Mode: $BUILD_MODE"
if [[ "$USE_FLIGHT_MOCK" == "1" ]]; then
  echo "Flight service: mock"
else
  echo "Flight service: real"
fi
echo
echo "Endpoints:"
echo "  mcp-time    http://localhost:8081/sse"
echo "  mcp-flight  http://localhost:8082/sse"
echo "  mcp-chatbox http://localhost:8083/sse"
echo "  mcp-photo   http://localhost:8084/sse"
echo
echo "Logs:"
echo "  $LOG_DIR/mcp-time.log"
echo "  $LOG_DIR/mcp-flight.log"
echo "  $LOG_DIR/mcp-chatbox.log"
echo "  $LOG_DIR/mcp-photo.log"
echo
echo "Press Ctrl+C to stop all services."

wait

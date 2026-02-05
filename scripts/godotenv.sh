#!/bin/bash
# Wrapper script to find and execute godotenv
GODOTENV=$(which godotenv 2>/dev/null || echo "$HOME/go/bin/godotenv")
exec "$GODOTENV" "$@"


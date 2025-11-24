#!/usr/bin/env bash
set -euo pipefail

# Reset the local 'dis' database for development.
# This script terminates active connections and recreates the database owned by dis_user.

echo "Terminating active connections to 'dis'..."
sudo -u postgres psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='dis';"

echo "Dropping database 'dis' (if exists)..."
sudo -u postgres psql -c "DROP DATABASE IF EXISTS dis;"

echo "Creating database 'dis' owned by dis_user..."
sudo -u postgres psql -c "CREATE DATABASE dis OWNER dis_user;"

echo "Reset complete."

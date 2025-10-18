#!/bin/bash

set -e

MNGR_DIR="$HOME/.mngr"
CONFIG_FILE="$MNGR_DIR/config.yaml"
DB_FILE="$MNGR_DIR/main.db"
SCHEMA_FILE="$(dirname "$0")/db/schema.sql"

echo "Setting up Password Manager..."

if [ ! -d "$MNGR_DIR" ]; then
    echo "Creating directory: $MNGR_DIR"
    mkdir -p "$MNGR_DIR"
else
    echo "Directory already exists: $MNGR_DIR"
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating config file: $CONFIG_FILE"
    cat > "$CONFIG_FILE" << EOF
db_path: ~/.mngr/main.db
EOF
else
    echo "Config file already exists: $CONFIG_FILE"
fi

if [ ! -f "$DB_FILE" ]; then
    echo "Creating database: $DB_FILE"

    if [ ! -f "$SCHEMA_FILE" ]; then
        echo "Error: Schema file not found at $SCHEMA_FILE"
        exit 1
    fi

    if ! command -v sqlite3 &> /dev/null; then
        echo "Error: sqlite3 is not installed. Please install sqlite3 first."
        exit 1
    fi

    sqlite3 "$DB_FILE" < "$SCHEMA_FILE"
    echo "Database initialized with schema"
else
    echo "Database already exists: $DB_FILE"
fi

echo ""
echo "Setup complete!"
echo "Directory: $MNGR_DIR"
echo "Config:    $CONFIG_FILE"
echo "Database:  $DB_FILE"

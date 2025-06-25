#!/bin/bash
# Check if database name argument is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <db-name>"
    exit 1
fi
# Get the database name from the first argument
SOURCE_PG_DB="$1"

# Remove 'delete_' prefix from the database name for the dump file
CLEAN_DB_NAME="${SOURCE_PG_DB#delete_}"

# Load environment variables
source .env
# Use current directory for dump file with the cleaned database name as prefix
DUMP_FILE="${CLEAN_DB_NAME}.sql"
echo "Starting database dump process..."
echo "Source DB: ${SOURCE_PG_DB}"
echo "Output file will be: ${DUMP_FILE}"
# Create the dump file through SSH
ssh -i "$SOURCE_SSH_KEY" "$SOURCE_SSH_USER@$SOURCE_SSH_HOST" \
    "PGPASSWORD='$SOURCE_PG_PASSWORD' pg_dump -h $SOURCE_PG_HOST -p $SOURCE_PG_PORT -U $SOURCE_PG_USER -d $SOURCE_PG_DB -F p --no-owner --no-acl --no-privileges" > "$DUMP_FILE"
# Check if dump was successful
if [ ! -s "$DUMP_FILE" ]; then
    echo "Error: Database dump is empty"
    exit 1
fi
echo "Database dump completed successfully!"
echo "Dump file: $DUMP_FILE"
echo "File size: $(du -h "$DUMP_FILE" | cut -f1)"

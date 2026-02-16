#!/bin/bash
# Helper script to create new timestamp-based migration files
# Usage: ./scripts/new-migration.sh migration_name

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <migration_name>"
    echo "Example: $0 add_user_preferences"
    exit 1
fi

MIGRATION_NAME="$1"
TIMESTAMP=$(date +%Y%m%d%H%M%S)
MIGRATIONS_DIR="internal/database/migrations"

# Sanitize migration name (replace spaces with underscores, lowercase)
MIGRATION_NAME=$(echo "$MIGRATION_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')

UP_FILE="${MIGRATIONS_DIR}/${TIMESTAMP}_${MIGRATION_NAME}.up.sql"
DOWN_FILE="${MIGRATIONS_DIR}/${TIMESTAMP}_${MIGRATION_NAME}.down.sql"

# Create .up.sql file
cat > "$UP_FILE" <<EOF
-- Migration: ${MIGRATION_NAME}
-- Created: $(date)

-- Add your UP migration SQL here

EOF

# Create .down.sql file
cat > "$DOWN_FILE" <<EOF
-- Migration: ${MIGRATION_NAME}
-- Created: $(date)

-- Add your DOWN migration SQL here (rollback)

EOF

echo "✅ Created migration files:"
echo "   📄 $UP_FILE"
echo "   📄 $DOWN_FILE"
echo ""
echo "Next steps:"
echo "1. Edit the .up.sql file with your schema changes"
echo "2. Edit the .down.sql file with the rollback logic"
echo "3. Test locally before committing"
echo ""

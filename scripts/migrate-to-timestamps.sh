#!/bin/bash
# Script to rename sequential migrations to timestamp format
# IMPORTANT: Run the generated update-schema-migrations.sql on all databases after renaming!

set -e

MIGRATIONS_DIR="/mounts/applications/sources/industry-tool/internal/database/migrations"
SQL_SCRIPT="/mounts/applications/sources/industry-tool/scripts/update-schema-migrations.sql"

# Timestamp mapping (maintains chronological order)
declare -A VERSION_MAP=(
    [1]="20250101_100000"
    [2]="20250101_110000"
    [3]="20250101_120000"
    [4]="20250101_130000"
    [5]="20250101_140000"
    [6]="20250101_150000"
    [7]="20250101_160000"
    [8]="20250101_170000"
    [9]="20250101_180000"
    [10]="20250101_190000"
    [11]="20250101_200000"
    [12]="20250101_210000"
)

echo "Creating SQL script to update schema_migrations table..."
cat > "$SQL_SCRIPT" <<'EOF'
-- Run this SQL script on ALL databases (dev, staging, production)
-- AFTER renaming migration files to update schema_migrations table

BEGIN;

-- Update version numbers to new timestamp format
UPDATE schema_migrations SET version = 20250101100000 WHERE version = 1;
UPDATE schema_migrations SET version = 20250101110000 WHERE version = 2;
UPDATE schema_migrations SET version = 20250101120000 WHERE version = 3;
UPDATE schema_migrations SET version = 20250101130000 WHERE version = 4;
UPDATE schema_migrations SET version = 20250101140000 WHERE version = 5;
UPDATE schema_migrations SET version = 20250101150000 WHERE version = 6;
UPDATE schema_migrations SET version = 20250101160000 WHERE version = 7;
UPDATE schema_migrations SET version = 20250101170000 WHERE version = 8;
UPDATE schema_migrations SET version = 20250101180000 WHERE version = 9;
UPDATE schema_migrations SET version = 20250101190000 WHERE version = 10;
UPDATE schema_migrations SET version = 20250101200000 WHERE version = 11;
UPDATE schema_migrations SET version = 20250101210000 WHERE version = 12;

-- Verify updates
SELECT version FROM schema_migrations ORDER BY version;

COMMIT;
EOF

echo "SQL script created at: $SQL_SCRIPT"
echo ""
echo "Renaming migration files..."

cd "$MIGRATIONS_DIR"

# Rename each migration file
for old_version in {1..12}; do
    new_timestamp="${VERSION_MAP[$old_version]}"

    for file in ${old_version}_*.sql; do
        if [ -f "$file" ]; then
            # Extract name and type (.up.sql or .down.sql)
            name_part=$(echo "$file" | sed "s/^${old_version}_//")
            new_name="${new_timestamp}_${name_part}"

            echo "Renaming: $file -> $new_name"
            mv "$file" "$new_name"
        fi
    done
done

echo ""
echo "✅ Migration files renamed successfully!"
echo ""
echo "⚠️  IMPORTANT NEXT STEPS:"
echo "1. Review the renamed files in: $MIGRATIONS_DIR"
echo "2. Run this SQL on ALL databases (dev, staging, production):"
echo "   psql -h localhost -p 19236 -U postgres -d app -f $SQL_SCRIPT"
echo "3. Verify schema_migrations table has new version numbers"
echo "4. Commit the renamed migration files"
echo ""

#!/bin/bash
# Create copilot_user for AI agent database access

echo "🤖 Creating copilot_user for AI agent access"

# Generated password
COPILOT_PASSWORD="aUE74gtSNhagLmySr1z1"

# Create user with read/write permissions
sudo -u postgres psql -d dis << EOF
-- Create copilot_user role
CREATE USER copilot_user WITH PASSWORD '$COPILOT_PASSWORD';

-- Grant connect to database
GRANT CONNECT ON DATABASE dis TO copilot_user;

-- Grant schema usage
GRANT USAGE ON SCHEMA public TO copilot_user;

-- Grant read/write on all existing tables
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO copilot_user;

-- Grant read/write on all future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO copilot_user;

-- Grant sequence usage (for serial/identity columns)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO copilot_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO copilot_user;

-- Verify creation
SELECT rolname, rolcanlogin FROM pg_roles WHERE rolname = 'copilot_user';
EOF

echo "✅ copilot_user created"

# Update .pgpass file
PGPASS_FILE="$HOME/.pgpass"

echo "📝 Updating $PGPASS_FILE"

# Backup existing .pgpass
cp "$PGPASS_FILE" "$PGPASS_FILE.backup.$(date +%s)"

# Remove old dis_user entry and add copilot_user
grep -v "localhost:5432:dis:dis_user:" "$PGPASS_FILE" > "$PGPASS_FILE.tmp" 2>/dev/null || touch "$PGPASS_FILE.tmp"
grep -v "localhost:5432:dis:copilot_user:" "$PGPASS_FILE.tmp" > "$PGPASS_FILE.tmp2" 2>/dev/null || touch "$PGPASS_FILE.tmp2"

# Add copilot_user entry
echo "localhost:5432:dis:copilot_user:$COPILOT_PASSWORD" >> "$PGPASS_FILE.tmp2"

# Move temp file to .pgpass
mv "$PGPASS_FILE.tmp2" "$PGPASS_FILE"
rm -f "$PGPASS_FILE.tmp"

# Set correct permissions
chmod 600 "$PGPASS_FILE"

echo "✅ .pgpass updated"

# Update .env.postgres
ENV_FILE="/home/rick/dev/DIS/dis-core/.env.postgres"

cat > "$ENV_FILE" << 'ENVEOF'
# PostgreSQL connection settings for DIS-Core
# Source this file: source .env.postgres

# AI Agent credentials (copilot_user)
export DIS_DB_DSN="postgres://copilot_user:aUE74gtSNhagLmySr1z1@localhost:5432/dis?sslmode=disable"

# Component settings (psql uses these)
export PGHOST=localhost
export PGPORT=5432
export PGDATABASE=dis
export PGUSER=copilot_user
export PGPASSWORD=aUE74gtSNhagLmySr1z1

# Disable pager to avoid "press q to exit" prompts
export PAGER=cat
export PSQL_PAGER=cat

# Alternative single-variable format
export DATABASE_URL="postgres://copilot_user:aUE74gtSNhagLmySr1z1@localhost:5432/dis?sslmode=disable"

echo "✅ PostgreSQL environment variables loaded (copilot_user)"
ENVEOF

echo "✅ .env.postgres updated with copilot_user credentials"

# Test connection
echo ""
echo "🧪 Testing copilot_user connection..."
source "$ENV_FILE"

if psql -c "SELECT current_user, current_database();" 2>&1 | grep -q "copilot_user"; then
    echo "✅ copilot_user connection successful!"
    echo ""
    echo "📋 Connection details:"
    echo "   User: copilot_user"
    echo "   Password: $COPILOT_PASSWORD"
    echo "   Database: dis"
    echo "   Permissions: SELECT, INSERT, UPDATE, DELETE on all tables"
else
    echo "⚠️  Connection test failed"
fi

echo ""
echo "🔄 To activate in current shell:"
echo "   source /home/rick/dev/DIS/dis-core/.env.postgres"

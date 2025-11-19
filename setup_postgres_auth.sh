#!/bin/bash
# Setup PostgreSQL authentication for dis-core development

echo "🔧 Setting up PostgreSQL authentication for DIS-Core"

# 1. Create .pgpass file for psql (avoids password prompts)
PGPASS_FILE="$HOME/.pgpass"

echo "📝 Creating/updating $PGPASS_FILE"

# Backup existing .pgpass if it exists
if [ -f "$PGPASS_FILE" ]; then
    cp "$PGPASS_FILE" "$PGPASS_FILE.backup.$(date +%s)"
    echo "   Backed up existing .pgpass"
fi

# Add dis-core entry (format: hostname:port:database:username:password)
# Remove any existing dis entries first
if [ -f "$PGPASS_FILE" ]; then
    grep -v "localhost:5432:dis:" "$PGPASS_FILE" > "$PGPASS_FILE.tmp" || true
    mv "$PGPASS_FILE.tmp" "$PGPASS_FILE"
fi

# Add the dis-core entry
echo "localhost:5432:dis:dis_user:card567" >> "$PGPASS_FILE"

# Set correct permissions (required by psql)
chmod 600 "$PGPASS_FILE"

echo "✅ .pgpass file configured"

# 2. Create environment variables file
ENV_FILE=".env.postgres"

cat > "$ENV_FILE" << 'EOF'
# PostgreSQL connection settings for DIS-Core
# Source this file: source .env.postgres

# Primary DSN (Go apps use this)
export DIS_DB_DSN="postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"

# Component settings (psql uses these)
export PGHOST=localhost
export PGPORT=5432
export PGDATABASE=dis
export PGUSER=dis_user
export PGPASSWORD=card567

# Alternative single-variable format
export DATABASE_URL="postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"

echo "✅ PostgreSQL environment variables loaded"
EOF

echo "✅ Created $ENV_FILE"

# 3. Add to .bashrc if not already there
BASHRC="$HOME/.bashrc"
if ! grep -q ".env.postgres" "$BASHRC" 2>/dev/null; then
    echo "" >> "$BASHRC"
    echo "# Auto-load DIS-Core PostgreSQL settings" >> "$BASHRC"
    echo "[ -f ~/dev/DIS/dis-core/.env.postgres ] && source ~/dev/DIS/dis-core/.env.postgres" >> "$BASHRC"
    echo "✅ Added auto-load to $BASHRC"
fi

# 4. Source the environment file now
source "$ENV_FILE"

# 5. Test connection
echo ""
echo "🧪 Testing PostgreSQL connection..."
# Use --no-psqlrc to avoid pager settings
if PAGER=cat psql --no-psqlrc -c "SELECT 'Connection successful!' as status;" 2>/dev/null | grep -q "Connection successful"; then
    echo "✅ PostgreSQL connection working!"
else
    echo "⚠️  Connection test failed. Please verify:"
    echo "   - PostgreSQL is running: sudo systemctl status postgresql"
    echo "   - User exists: sudo -u postgres psql -c \"\\du\""
    echo "   - Database exists: psql -l"
fi

echo ""
echo "📋 Summary:"
echo "   1. ~/.pgpass file created (no more password prompts for psql)"
echo "   2. .env.postgres file created in this directory"
echo "   3. Auto-load added to ~/.bashrc"
echo ""
echo "🔄 To use immediately in current shell:"
echo "   source .env.postgres"
echo ""
echo "🔄 Or restart your terminal for automatic loading"

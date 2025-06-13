#!/bin/bash

# Dutch Learning Bot Runner Script
# This script builds and runs the Dutch learning Telegram bot

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check dependencies
print_info "Checking dependencies..."

if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d'.' -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d'.' -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
    print_error "Go version $GO_VERSION detected. Please upgrade to Go 1.21 or later."
    exit 1
fi

print_success "Go version $GO_VERSION detected"

# Load .env file first if it exists
if [ -f ".env" ]; then
    print_info "Loading environment variables from .env file..."
    set -a  # Automatically export all variables
    source .env
    set +a
    print_success "Environment variables loaded from .env"
fi

# Check for bot token after loading .env
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    print_warning "TELEGRAM_BOT_TOKEN environment variable is not set."
    echo
    echo "To get a bot token:"
    echo "1. Message @BotFather on Telegram"
    echo "2. Send /newbot and follow instructions"
    echo "3. Copy your bot token"
    echo
    read -p "Enter your Telegram bot token: " BOT_TOKEN
    if [ -z "$BOT_TOKEN" ]; then
        print_error "Bot token is required. Exiting."
        exit 1
    fi
    
    # Export the token for this session
    export TELEGRAM_BOT_TOKEN="$BOT_TOKEN"
    print_success "Bot token set in environment for this session"
    
    # Ask if user wants to save token to .env file
    echo
    read -p "Save token to .env file for future runs? (y/N): " SAVE_TOKEN
    if [[ $SAVE_TOKEN =~ ^[Yy]$ ]]; then
        echo "TELEGRAM_BOT_TOKEN=$BOT_TOKEN" > .env
        print_success "Token saved to .env file"
        echo "Future runs will automatically load this token"
    fi
else
    print_success "Bot token found in environment"
fi

# Validate that bot token is actually set and looks valid
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    print_error "Bot token is still not set. Cannot proceed."
    exit 1
fi

# Basic token format validation (should contain colon and be reasonably long)
if [[ ! "$TELEGRAM_BOT_TOKEN" =~ ^[0-9]+:[A-Za-z0-9_-]+$ ]] || [ ${#TELEGRAM_BOT_TOKEN} -lt 30 ]; then
    print_error "Bot token format appears invalid. Expected format: 123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
    print_error "Current token: ${TELEGRAM_BOT_TOKEN:0:20}..."
    exit 1
fi

print_success "Bot token validated and ready to use"

# Create logs directory if it doesn't exist
mkdir -p logs

# Build the bot
print_info "Building the Dutch Learning Bot..."
if go build -o bot ./cmd/bot; then
    print_success "Bot built successfully"
else
    print_error "Failed to build bot"
    exit 1
fi

# Database and migration handling
DB_NAME="dutch_learning.db"
MIGRATION_FILE="migrate_categories.sql"

# Function to run database migration
run_migration() {
    if [ ! -f "$MIGRATION_FILE" ]; then
        print_warning "Migration file $MIGRATION_FILE not found - skipping migration"
        return 0
    fi
    
    if [ ! -f "$DB_NAME" ]; then
        print_info "Database doesn't exist yet - will be created on first run"
        return 0
    fi
    
    # Check if tables exist
    TABLES_EXIST=$(sqlite3 "$DB_NAME" "SELECT name FROM sqlite_master WHERE type='table' AND name='words';" 2>/dev/null || echo "")
    if [ -z "$TABLES_EXIST" ]; then
        print_info "Database tables don't exist yet - migration will run after initial setup"
        return 0
    fi
    
    print_info "Checking if database migration is needed..."
    
    # Backup existing database
    BACKUP_NAME="${DB_NAME}.backup.$(date +%Y%m%d_%H%M%S)"
    print_info "Creating database backup: $BACKUP_NAME"
    cp "$DB_NAME" "$BACKUP_NAME"
    
    # Check if migration is needed (look for old categories)
    OLD_CATEGORIES=$(sqlite3 "$DB_NAME" "SELECT COUNT(*) FROM words WHERE category IN ('particles', 'verbs', 'common_verbs') LIMIT 1;" 2>/dev/null || echo "0")
    
    if [ "$OLD_CATEGORIES" -gt 0 ]; then
        print_info "Old categories detected. Running migration..."
        if sqlite3 "$DB_NAME" < "$MIGRATION_FILE"; then
            print_success "Database migration completed successfully"
            
            # Show new category distribution
            print_info "New category distribution:"
            sqlite3 "$DB_NAME" "SELECT category, COUNT(*) as count FROM words GROUP BY category ORDER BY count DESC;" 2>/dev/null || true
        else
            print_error "Migration failed. Database backup preserved."
            exit 1
        fi
    else
        print_info "No migration needed - database already up to date"
        # Remove backup since no changes were made
        rm "$BACKUP_NAME" 2>/dev/null || true
    fi
}

# Function to run delayed migration (after bot creates tables)
run_delayed_migration() {
    if [ -f "$MIGRATION_FILE" ] && [ -f "$DB_NAME" ]; then
        # Wait a bit for bot to create tables and load initial data
        sleep 5
        
        # Check if tables now exist and have old categories
        TABLES_EXIST=$(sqlite3 "$DB_NAME" "SELECT name FROM sqlite_master WHERE type='table' AND name='words';" 2>/dev/null || echo "")
        if [ -n "$TABLES_EXIST" ]; then
            OLD_CATEGORIES=$(sqlite3 "$DB_NAME" "SELECT COUNT(*) FROM words WHERE category IN ('particles', 'verbs', 'common_verbs') LIMIT 1;" 2>/dev/null || echo "0")
            
            if [ "$OLD_CATEGORIES" -gt 0 ]; then
                print_info "Running delayed migration after database initialization..."
                
                # Backup before migration
                BACKUP_NAME="${DB_NAME}.backup.delayed.$(date +%Y%m%d_%H%M%S)"
                cp "$DB_NAME" "$BACKUP_NAME"
                
                if sqlite3 "$DB_NAME" < "$MIGRATION_FILE"; then
                    print_success "Delayed migration completed successfully"
                    print_info "New category distribution:"
                    sqlite3 "$DB_NAME" "SELECT category, COUNT(*) as count FROM words GROUP BY category ORDER BY count DESC;" 2>/dev/null || true
                else
                    print_error "Delayed migration failed"
                fi
            fi
        fi
    fi
}

# Run migration before starting the bot
run_migration

# Check if vocabulary file exists
if [ ! -f "vocabulary.json" ]; then
    print_warning "vocabulary.json not found. The bot may not work properly."
    echo "Make sure you have a vocabulary.json file with Dutch words."
fi

# Function to run the bot with restart capability
run_bot() {
    local restart_count=0
    local max_restarts=5
    local restart_delay=5
    local migration_attempted=false

    while [ $restart_count -lt $max_restarts ]; do
        print_info "Starting Dutch Learning Bot... (attempt $((restart_count + 1)))"
        echo "$(date): Starting bot (attempt $((restart_count + 1)))" >> logs/bot.log
        
        # Start bot in background for migration check
        ./bot 2>&1 | tee -a logs/bot.log &
        local bot_pid=$!
        
        # Run delayed migration after a short delay (only on first attempt)
        if [ "$restart_count" -eq 0 ] && [ "$migration_attempted" = false ]; then
            migration_attempted=true
            (
                sleep 10  # Give bot time to create database and load data
                run_delayed_migration
            ) &
        fi
        
        # Wait for bot to finish
        if wait $bot_pid; then
            print_success "Bot exited normally"
            break
        else
            exit_code=$?
            restart_count=$((restart_count + 1))
            
            if [ $restart_count -lt $max_restarts ]; then
                print_warning "Bot crashed with exit code $exit_code. Restarting in ${restart_delay}s..."
                echo "$(date): Bot crashed with exit code $exit_code. Restarting..." >> logs/bot.log
                sleep $restart_delay
                restart_delay=$((restart_delay * 2))  # Exponential backoff
            else
                print_error "Bot crashed $max_restarts times. Giving up."
                echo "$(date): Bot crashed $max_restarts times. Giving up." >> logs/bot.log
                exit $exit_code
            fi
        fi
    done
}

# Handle cleanup on exit
cleanup() {
    print_info "Shutting down bot..."
    jobs -p | xargs -r kill
    exit 0
}

trap cleanup SIGINT SIGTERM

# Show startup information
echo
print_success "=== Dutch Learning Bot ==="
print_info "Bot token: ${TELEGRAM_BOT_TOKEN:0:10}... (${#TELEGRAM_BOT_TOKEN} chars)"
print_info "Database: dutch_learning.db"
print_info "Logs: logs/bot.log"
print_info "Smart reminders: ENABLED"
print_info "Commands: Will be auto-configured with Telegram"
print_info "Press Ctrl+C to stop the bot"

# Debug: Confirm environment is properly set
print_info "Environment check: TELEGRAM_BOT_TOKEN is $([ -n "$TELEGRAM_BOT_TOKEN" ] && echo "SET" || echo "NOT SET")"
echo

# Ask if user wants to enable auto-restart
read -p "Enable auto-restart on crashes? (Y/n): " AUTO_RESTART
if [[ ! $AUTO_RESTART =~ ^[Nn]$ ]]; then
    print_info "Auto-restart enabled (max 5 attempts)"
    run_bot
else
    print_info "Auto-restart disabled"
    echo "$(date): Starting bot (no auto-restart)" >> logs/bot.log
    ./bot 2>&1 | tee -a logs/bot.log
fi

print_success "Bot stopped" 
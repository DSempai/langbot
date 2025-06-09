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

# Check if database exists, if not, it will be created automatically
if [ ! -f "dutch_learning.db" ]; then
    print_info "Database will be created on first run"
fi

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

    while [ $restart_count -lt $max_restarts ]; do
        print_info "Starting Dutch Learning Bot... (attempt $((restart_count + 1)))"
        echo "$(date): Starting bot (attempt $((restart_count + 1)))" >> logs/bot.log
        
        # Run the bot and capture exit code
        if ./bot 2>&1 | tee -a logs/bot.log; then
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
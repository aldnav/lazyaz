#!/bin/bash

# Debug logging function
debug_log() {
    if [ "$LAZYAZ_DEBUG" = "1" ]; then
        echo "$@" | tee -a "$LOG_FILE"
    else
        echo "$@" >> "$LOG_FILE"
    fi
}

# Create logs directory if it doesn't exist
mkdir -p logs

# Set log file with timestamp
LOG_FILE="logs/debug_$(date +%Y%m%d_%H%M%S).log"
debug_log "Debug log will be saved to: $LOG_FILE"

# Check if .env file exists and source it
if [ -f .env ]; then
    debug_log "Loading environment variables from .env file"
    source .env
fi

# Set your Azure DevOps organization
# This value will be used if not set in .env file or Azure CLI config
if [ -z "$AZURE_DEVOPS_ORG" ]; then
    export AZURE_DEVOPS_ORG="your-organization-name"
fi

# Check if Azure CLI is installed
if ! command -v az &> /dev/null; then
    echo "Azure CLI is not installed. Please install it first:"
    echo "https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
fi

# Check if Azure CLI is logged in
debug_log "Checking Azure CLI login status..."
if ! az account show &> /dev/null; then
    echo "You are not logged in to Azure CLI. Please run 'az login' first."
    exit 1
fi

# Check if Azure DevOps extension is installed
if ! az extension list --query "[].{name:name} | [].name" | grep "azure-devops" &> /dev/null; then
    debug_log "Azure DevOps CLI extension is not installed. Installing it now..."
    az extension add --name azure-devops
fi

# Try to capture the organization from the Azure CLI again
if [ "$AZURE_DEVOPS_ORG" == "your-organization-name" ]; then
    AZURE_DEVOPS_ORG=$(az devops configure --list | grep "organization" | awk '{print $3}')
fi

# Run the application
debug_log "---------------------------------------------"
debug_log "Starting LazyAZ application at $(date)"
debug_log "Using Azure DevOps Organization: $AZURE_DEVOPS_ORG"
debug_log "---------------------------------------------"

# Run application and capture output to log file
go run ./app 2>&1 | tee -a "$LOG_FILE"

# Alternatively, if you've built the application using 'go build' or 'make build'
# uncomment the line below and comment out the 'go run' line above
# ./lazyaz 2>&1 | tee -a "$LOG_FILE"

debug_log "---------------------------------------------"
debug_log "LazyAZ application exited at $(date)"
debug_log "Exit code: $?"
debug_log "Debug log saved to: $LOG_FILE"
debug_log "---------------------------------------------"

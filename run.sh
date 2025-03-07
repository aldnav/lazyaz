#!/bin/bash

# Create logs directory if it doesn't exist
mkdir -p logs

# Set log file with timestamp
LOG_FILE="logs/debug_$(date +%Y%m%d_%H%M%S).log"
echo "Debug log will be saved to: $LOG_FILE"

# Check if .env file exists and source it
if [ -f .env ]; then
    echo "Loading environment variables from .env file"
    source .env
fi

# Set your Azure DevOps organization and personal access token
# These values will be used if not set in .env file
if [ -z "$AZURE_DEVOPS_ORG" ]; then
    export AZURE_DEVOPS_ORG="your-organization-name"
fi

if [ -z "$AZURE_DEVOPS_TOKEN" ]; then
    export AZURE_DEVOPS_TOKEN="your-personal-access-token"
fi

# Run the application
echo "---------------------------------------------" | tee -a "$LOG_FILE"
echo "Starting LazyAZ application at $(date)" | tee -a "$LOG_FILE"
echo "Using Azure DevOps Organization: $AZURE_DEVOPS_ORG" | tee -a "$LOG_FILE"
echo "Token is ${#AZURE_DEVOPS_TOKEN} characters long" | tee -a "$LOG_FILE"
echo "---------------------------------------------" | tee -a "$LOG_FILE"

# Run application and capture output to log file
go run main.go 2>&1 | tee -a "$LOG_FILE"

# Alternatively, if you've built the application using 'go build' or 'make build'
# uncomment the line below and comment out the 'go run' line above
# ./lazyaz 2>&1 | tee -a "$LOG_FILE"

echo "---------------------------------------------" | tee -a "$LOG_FILE"
echo "LazyAZ application exited at $(date)" | tee -a "$LOG_FILE"
echo "Exit code: $?" | tee -a "$LOG_FILE"
echo "Debug log saved to: $LOG_FILE" | tee -a "$LOG_FILE"
echo "---------------------------------------------" | tee -a "$LOG_FILE"

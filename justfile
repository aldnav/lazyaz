# Set variables
binary_name := "lazyaz"
log_dir := "logs"

# Default recipe when just is called without arguments
default:
    @just --list

# Build the application
build:
    go build -o bin/{{binary_name}} -v ./app

# Run all tests
test:
    go test -v ./...

# Run with specific test pattern
test-filter pattern:
    go test -v ./... -run "{{pattern}}"

# Run the application using run.sh
run: build
    ./run.sh

# Run the application directly (alternative to run.sh)
run-direct:
    #!/usr/bin/env bash
    # Debug logging function
    debug_log() {
        if [ "$LAZYAZ_DEBUG" = "1" ]; then
            echo "$@" | tee -a "$LOG_FILE"
        else
            echo "$@" >> "$LOG_FILE"
        fi
    }

    # Create logs directory if it doesnt exist
    mkdir -p {{log_dir}}
    
    # Set log file with timestamp
    LOG_FILE="{{log_dir}}/debug_$(date +%Y%m%d_%H%M%S).log"
    debug_log "Debug log will be saved to: $LOG_FILE"
    
    # Check if .env file exists and source it
    if [ -f .env ]; then
        debug_log "Loading environment variables from .env file"
        source .env
    fi
    
    # Run application and capture output to log file
    debug_log "---------------------------------------------"
    debug_log "Starting LazyAZ application at $(date)"
    debug_log "Using Azure DevOps Organization: $AZURE_DEVOPS_ORG"
    debug_log "---------------------------------------------"
    
    # Run the binary if built, otherwise use go run
    if [ -f "bin/{{binary_name}}" ]; then
        ./bin/{{binary_name}} 2>&1 | tee -a "$LOG_FILE"
    else
        go run ./app 2>&1 | tee -a "$LOG_FILE"
    fi
    
    debug_log "---------------------------------------------"
    debug_log "LazyAZ application exited at $(date)"
    debug_log "Exit code: $?"
    debug_log "Debug log saved to: $LOG_FILE"
    debug_log "---------------------------------------------"

# Clean build artifacts
clean:
    go clean
    rm -rf bin/
    rm -rf dist/
    rm -f {{binary_name}}

# Clean logs
clean-logs:
    rm -rf {{log_dir}}

# Clean everything
clean-all: clean clean-logs

# Read Azure DevOps organization from config file
read-azure-org:
    #!/usr/bin/env bash
    # Try azuredevops config first
    ORG=$(grep "organization =" ~/.azure/azuredevops/config | awk -F "= " '{print $2}')
    if [ -z "$ORG" ]; then
        # If empty, try main azure config
        ORG=$(grep "^organization =" ~/.azure/config | awk -F "= " '{print $2}')
    fi
    echo "Azure DevOps Organization from config: $ORG"

release-healthcheck:
    goreleaser healthcheck

release-build:
    goreleaser build --snapshot --clean

release-snapshot:
    goreleaser release --snapshot --skip=publish --clean

release-dryrun: release-healthcheck release-build
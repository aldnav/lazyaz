# Set variables
binary_name := "lazyaz"
log_dir := "logs"

# Default recipe when just is called without arguments
default:
    @just --list

# Build the application
build:
    go build -o bin/{{binary_name}} -v .

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
    # Create logs directory if it doesn't exist
    mkdir -p {{log_dir}}
    
    # Set log file with timestamp
    LOG_FILE="{{log_dir}}/debug_$(date +%Y%m%d_%H%M%S).log"
    echo "Debug log will be saved to: $LOG_FILE"
    
    # Check if .env file exists and source it
    if [ -f .env ]; then
        echo "Loading environment variables from .env file"
        source .env
    fi
    
    # Run application and capture output to log file
    echo "---------------------------------------------" | tee -a "$LOG_FILE"
    echo "Starting LazyAZ application at $(date)" | tee -a "$LOG_FILE"
    echo "Using Azure DevOps Organization: $AZURE_DEVOPS_ORG" | tee -a "$LOG_FILE"
    echo "Token is ${#AZURE_DEVOPS_TOKEN} characters long" | tee -a "$LOG_FILE"
    echo "---------------------------------------------" | tee -a "$LOG_FILE"
    
    # Run the binary if built, otherwise use go run
    if [ -f "bin/{{binary_name}}" ]; then
        ./bin/{{binary_name}} 2>&1 | tee -a "$LOG_FILE"
    else
        go run main.go 2>&1 | tee -a "$LOG_FILE"
    fi
    
    echo "---------------------------------------------" | tee -a "$LOG_FILE"
    echo "LazyAZ application exited at $(date)" | tee -a "$LOG_FILE"
    echo "Exit code: $?" | tee -a "$LOG_FILE"
    echo "Debug log saved to: $LOG_FILE" | tee -a "$LOG_FILE"
    echo "---------------------------------------------" | tee -a "$LOG_FILE"

# Clean build artifacts
clean:
    go clean
    rm -rf bin/
    rm -f {{binary_name}}

# Clean logs
clean-logs:
    rm -rf {{log_dir}}

# Clean everything
clean-all: clean clean-logs

# Read Azure DevOps organization from config file
read-azure-org:
    #!/usr/bin/env bash
    ORG=$(grep "organization =" ~/.azure/azuredevops/config | awk -F "= " '{print $2}')
    echo "Azure DevOps Organization from config: $ORG"


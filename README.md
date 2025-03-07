# LazyAZ

A terminal-based UI application for interacting with Azure DevOps, built with Go.

## Overview

LazyAZ provides a convenient terminal interface to interact with Azure DevOps services. It allows you to:

- View Azure DevOps projects
- Navigate through project resources 
- Perform common Azure DevOps operations from your terminal

## Prerequisites

- Go 1.16 or higher
- Azure DevOps account
- Azure CLI installed with the Azure DevOps extension

## Setup

1. Clone the repository:
   ```
   git clone https://github.com/aldnav/lazyaz.git
   cd lazyaz
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Authentication

LazyAZ now uses Azure CLI for authentication instead of a personal access token. Make sure you have logged in with Azure CLI before using the application:

```bash
az login
az devops configure --defaults organization=https://dev.azure.com/your-organization
```

You can also set a default project in the Azure DevOps configuration file:

```
[defaults]
organization = your-org-name
project = your-project-name
```

This configuration is stored in `~/.azure/azuredevops/config`.

### Environment Variables

While the application now primarily uses Azure CLI for authentication, you can still set the following environment variable:

| Variable | Description | Required |
|----------|-------------|----------|
| AZURE_DEVOPS_ORG | Your Azure DevOps organization name | No (if configured in Azure CLI) |
| AZURE_DEVOPS_PROJECT | Your default Azure DevOps project | No (if configured in Azure CLI) |

You can set these environment variables in your shell:

```bash
export AZURE_DEVOPS_ORG="your-organization"
export AZURE_DEVOPS_PROJECT="your-project"
```

Or create a `.env` file (make sure to add to `.gitignore`) and load it before running the application.

For convenience, a `.env.example` file is provided in the repository. You can copy this file to create your own `.env`:

```bash
cp .env.example .env
# Edit the .env file with your actual values
```

## Build

To build the application:

```bash
go build -o bin/lazyaz
```

Or using just:

```bash
just build
```

## Run

To run the application directly:

```bash
go run main.go
```

To run the built binary:

```bash
./bin/lazyaz
```

Or using just:

```bash
just run
```

Using the run.sh script (recommended for environment variables setup):

```bash
./run.sh
```

This script allows you to set your Azure DevOps organization in one place and run the application. The application uses Azure CLI for authentication, so make sure you're logged in with `az login` before running.

## Testing

To run tests:

```bash
go test ./...
```

Or using just:

```bash
just test
```

## Project Structure

- `cmd/lazyaz/`: Main application entry point
- `pkg/`: Reusable libraries
- `internal/`: Internal packages not meant for external use

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.


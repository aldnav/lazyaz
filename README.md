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
- Personal Access Token (PAT) with appropriate permissions

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

## Environment Variables

LazyAZ requires the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| AZURE_DEVOPS_ORG | Your Azure DevOps organization name | Yes |
| AZURE_DEVOPS_TOKEN | Personal Access Token with appropriate permissions | Yes |

You can set these environment variables in your shell:

```bash
export AZURE_DEVOPS_ORG="your-organization"
export AZURE_DEVOPS_TOKEN="your-personal-access-token"
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

This script allows you to set your Azure DevOps organization and token in one place and run the application. Edit the script to add your own values.

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


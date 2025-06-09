# RuneBird Emailer Service

RuneBird is a self-hosted, containerized email service written in Go, designed to send templated HTML emails via a REST API.
It supports immediate and scheduled email delivery, global rate limiting, and Prometheus-compatible metrics for observability.
Perfect for integration into other applications using Docker Compose.

[![Go](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com)
![Github license](https://img.shields.io/github/license/martinsedd/runebird)

## Overview

RuneBird provides a lightweight, configurable solution for sending HTML emails with templated content. Whether you need to send transactional emails, newsletters, or scheduled notifications, RuneBird offers a robust API to handle your email needs with ease.

## Features

- **Configuration**: Fully configurable via a YAML file with environment variable overrides.
- **Email Sending**: Send HTML emails to multiple recipients using SMTP.
- **Templating**: Render emails with Go's `html/template` engine, supporting logic and custom subject lines.
- **Rate Limiting**: Enforce global send limits with delayed retries to prevent drops.
- **Scheduling**: Schedule emails for future delivery in UTC.
- **Observability**: Structured logging (stdout/stderr and file) and Prometheus metrics at `/metrics`.
- **Deployment**: Runs as a single Docker container, easily integrable with Docker Compose.

## Getting Started

### Prerequisites

- **Go 1.24+**: Required for local development and building.
- **Docker & Docker Compose**: For containerized deployment (recommended for production).
- **SMTP Credentials**: Necessary for sending emails (configure in `emailer.yaml`).

### Installation

### Using Docker (Recommended for Production)

1. Clone the repository:
    ```bash
    git clone https://github.com/martinsedd/runebird.git
    cd runebird
    ```
2. Configure `emailer.yaml` with your SMTP credentials and desired settings.
3. Build and run the container:
    ```bash
    docker-compose up -d
    ```
4. Access the API at `http://localhost:8080`.

#### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/<your-username>/runebird.git
   cd runebird
   ```

2. Ensure the `./templates` directory contains at least one `.html` template file:
   ```bash
   mkdir -p templates
   echo "<html><body><p>Placeholder template for RuneBird.</p></body></html>" > templates/placeholder.html
   ```

3. Configure `emailer.yaml` with your SMTP credentials.

4. Build and run the application:
   ```bash
   go build -o runebird ./cmd/emailer
   ./runebird
   ```

## API Endpoints

### Send Immediate Email (`/send`)

Send an email immediately to one or more recipients using a specified template.

```bash
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{
    "template": "welcome",
    "recipients": ["user@example.com"],
    "data": {
      "Name": "Alice"
    }
  }'
```

**Response**:
```json
{"status": "success"}
```

### Schedule Future Email (`/schedule`)

Schedule an email to be sent at a future time (UTC).

```bash
curl -X POST http://localhost:8080/schedule \
  -H "Content-Type: application/json" \
  -d '{
    "template": "digest",
    "recipients": ["user@example.com"],
    "send_at": "2025-06-10T15:00:00Z",
    "data": {
      "Day": "Tuesday"
    }
  }'
```

**Response**:
```json
{"status": "success", "task_id": "sched-1234567890123456"}
```

### Metrics (`/metrics`)

Access Prometheus-compatible metrics for monitoring.

```bash
curl http://localhost:8080/metrics
```

## Configuration

RuneBird is configured via `emailer.yaml`. Below is an example configuration:

```yaml
server:
  port: 8080
smtp:
  host: "smtp.example.com"
  port: 587
  username: "user@example.com"
  password: "your-smtp-password"
  from_address: "no-reply@runebird.app"
templates:
  path: "./templates"
rate_limit:
  per_hour: 100
  burst: 5
logging:
  file_path: "./logs/runebird.log"
  level: "info"
```

You can override the config file path with the `EMAILER_CONFIG_PATH` environment variable.

## Project Structure

```text
runebird/
├── cmd/emailer/            # Application entry point
├── internal/               # Core packages (config, email, templates, etc.)
│   ├── config/             # YAML configuration loading
│   ├── email/              # SMTP email sending logic
│   ├── templates/          # Templating engine
│   ├── server/             # HTTP API server
│   ├── rate/               # Rate limiting
│   ├── scheduler/          # Scheduled email handling
│   └── logger/             # Structured logging
├── templates/              # Directory for HTML email templates
├── logs/                   # Directory for log output
├── emailer.yaml            # Default configuration file
├── Dockerfile              # Docker image definition
└── docker-compose.yaml     # Docker Compose configuration
```

## Roadmap

RuneBird is actively developed. Here are some of the features and improvements planned for future releases (v2+). Contributions are welcome!

-   **Enhanced API Security**: Implement an API key and OAuth-based authentication.
-   **Web-based Admin UI**: A simple web interface to preview templates, monitor sends, and view logs.
-   **Multiple SMTP Backends**: Support for multiple SMTP providers with failover logic.
-   **Markdown Template Support**: Write email templates in Markdown and have them automatically converted to HTML.
-   **Inbound Email Processing**: Handle incoming emails via IMAP/POP3 to enable reply tracking or other automations.
-   **Multi-Tenancy**: Support for isolated configurations for different users or applications.
- 
## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to report bugs, suggest features, or submit pull requests. All participants are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## License

This project is licensed under the MIT License—see the [LICENSE](LICENSE) file for details.

## Contact

For questions, support, or feedback, please open an issue on GitHub or contact the maintainer at [martins.edd04@gmail.com](mailto:martins.edd04@gmail.com).
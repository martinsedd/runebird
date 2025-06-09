# Contributing to RuneBird

Thank you for your interest in contributing to RuneBird, a self-hosted email service for sending templated HTML emails via REST API. This guide outlines the process for contributing to the project. By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How Can I Contribute?

There are many ways to contribute to RuneBird, whether you're a developer, designer, or just have ideas to share. Here are some options:

-   **Bug Reports**: If you encounter a bug, please open an issue with a detailed description, steps to reproduce, expected behavior, and any relevant logs or screenshots.
-   **Feature Requests**: Have an idea for a new feature or improvement? Open an issue with the label "feature request" and describe your proposal.
-   **Code Contributions**: Want to fix a bug or implement a feature? Follow the steps below to submit a pull request.
-   **Documentation**: Improve project documentation, README, API guides, or examples by submitting updates or corrections.
-   **Testing**: Help by writing or enhancing test cases to ensure the reliability of the service.

## Development Setup

To set up a local development environment for RuneBird:

1.  **Clone the Repository**:
    ```bash
    git clone https://github.com/<your-username>/runebird.git
    cd runebird
    ```

2.  **Install Dependencies**:
    Ensure you have Go 1.22+ installed, then download dependencies:
    ```bash
    go mod download
    ```

3.  **Create a Template**:
    Add at least one `.html` file in the `./templates` directory for testing email sends:
    ```bash
    mkdir -p templates
    echo "<html><body><p>Placeholder template for RuneBird.</p></body></html>" > templates/placeholder.html
    ```

4.  **Configure SMTP**:
    Update `emailer.yaml` with valid SMTP credentials for email sending.

5.  **Run the Application**:
    ```bash
    go run ./cmd/emailer
    ```
    The API will be available at `http://localhost:8080`.

6.  **Run Tests**:
    ```bash
    go test ./internal/... -v
    ```

## Code Contribution Guidelines

We strive to maintain a clean, maintainable codebase that adheres to Go best practices. Please follow these guidelines when contributing code:

-   **Code Style**: Use Go's standard formatting with `gofmt`. Run `go fmt ./...` before committing.
-   **Testing**: Write tests for new functionality or bug fixes. Ensure existing tests pass with `go test ./internal/...`.
-   **Commit Messages**: Write clear, descriptive commit messages starting with a capital letter. Reference related issues if applicable (e.g., "Fix template loading error (#123)").
-   **Pull Requests**: Submit pull requests to the `main` branch. Include a detailed description of changes, related issues, and testing performed.
-   **Dependencies**: Avoid adding unnecessary dependencies. Discuss significant additions in an issue before implementation.

## Pull Request Process

1.  **Create a Branch**: Create a branch with a descriptive name related to the feature or bug (e.g., `feature/add-oauth-support` or `bugfix/fix-rate-limiter`).
2.  **Make Changes**: Implement your changes, commit them with meaningful messages, and push to your branch.
3.  **Update Documentation**: If applicable, update `README.md` or other documentation to reflect your changes.
4.  **Run Tests**: Ensure all tests pass before submitting your pull request.
5.  **Submit PR**: Open a pull request against the `main` branch. Fill in the PR template (if available) with details of your changes.
6.  **Code Review**: Address any feedback or requested changes from maintainers promptly.

## Community

For questions, discussions, or help, feel free to open an issue on GitHub. We aim to respond to issues and pull requests within a reasonable timeframe.

Thank you for contributing to RuneBird and helping make it better!
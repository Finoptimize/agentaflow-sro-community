# Contributing to AgentaFlow SRO Community

We welcome contributions to AgentaFlow SRO! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/agentaflow-sro-community.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make test`
6. Commit your changes: `git commit -m "Add your feature"`
7. Push to your fork: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Building the Project

```bash
# Install dependencies
make deps

# Build the project
make build

# Run tests
make test

# Run tests with coverage
make test-coverage
```

## Code Style

We follow standard Go conventions:

- Use `go fmt` to format your code
- Use `go vet` to check for common mistakes
- Write meaningful commit messages
- Add comments for exported functions and types

Run code quality checks:

```bash
make check
```

## Testing

All new features should include tests:

- Unit tests for individual components
- Integration tests for component interactions
- Examples demonstrating usage

Test files should be named `*_test.go` and placed in the same package as the code they test.

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Project Structure

```
agentaflow-sro-community/
├── pkg/                  # Core packages
│   ├── gpu/             # GPU orchestration
│   ├── serving/         # Model serving
│   └── observability/   # Monitoring and debugging
├── cmd/                 # Command-line applications
│   └── agentaflow/     # Main CLI
├── examples/            # Usage examples
└── docs/                # Documentation
```

## Pull Request Guidelines

1. **Title**: Use a clear, descriptive title
2. **Description**: Explain what your PR does and why
3. **Tests**: Include tests for new features
4. **Documentation**: Update documentation as needed
5. **Code Quality**: Ensure all checks pass

### PR Checklist

- [ ] Code follows Go conventions
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] No security vulnerabilities introduced
- [ ] Commit messages are clear
- [ ] PR description is comprehensive

## Feature Requests

We welcome feature requests! Please:

1. Check existing issues first
2. Create a new issue with the "enhancement" label
3. Describe the feature and its use case
4. Explain how it aligns with project goals

## Bug Reports

When reporting bugs, please include:

1. Go version and OS
2. Steps to reproduce
3. Expected behavior
4. Actual behavior
5. Error messages or logs
6. Minimal code example if applicable

## Areas for Contribution

We especially welcome contributions in these areas:

### GPU Orchestration
- Kubernetes integration
- Multi-cloud support
- Real GPU metrics collection
- Advanced scheduling algorithms

### Model Serving
- More routing strategies
- Auto-scaling support
- Advanced batching techniques
- Model versioning

### Observability
- Prometheus integration
- Grafana dashboards
- OpenTelemetry support
- Cost prediction models

### Documentation
- Tutorials and guides
- API documentation
- Architecture diagrams
- Performance benchmarks

### Infrastructure
- CI/CD improvements
- Docker support
- Helm charts
- Deployment scripts

## Code Review Process

1. Automated checks must pass
2. At least one maintainer review required
3. Address feedback promptly
4. Maintain respectful communication

## Community Guidelines

- Be respectful and inclusive
- Help others learn
- Share knowledge
- Collaborate constructively

## Questions?

- Open an issue for questions
- Check existing documentation
- Review examples for guidance

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be acknowledged in:
- Release notes
- Contributors file
- Project documentation

Thank you for contributing to AgentaFlow SRO!

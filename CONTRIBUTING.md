# Contributing to ISP Config Terraform Provider

Thank you for your interest in contributing to the ISP Config Terraform Provider!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/ispconfig-terraform-provider.git`
3. Create a new branch: `git checkout -b feature/my-new-feature`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes: `git commit -am 'Add some feature'`
7. Push to the branch: `git push origin feature/my-new-feature`
8. Submit a pull request

## Development Requirements

- Go 1.21 or later
- Terraform 1.0 or later
- Access to an ISP Config instance with remote API enabled (for testing)

## Code Style

This project follows standard Go coding conventions:

- Run `go fmt` before committing
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Write clear, descriptive commit messages

## Testing

### Unit Tests

Run unit tests with:

```bash
go test -v ./...
```

### Acceptance Tests

Acceptance tests require a running ISP Config instance. Set the following environment variables:

```bash
export ISPCONFIG_HOST="your-host:8080"
export ISPCONFIG_USERNAME="your-username"
export ISPCONFIG_PASSWORD="your-password"
export TF_ACC=1
```

Then run:

```bash
go test -v ./internal/provider -timeout 120m
```

## Adding New Resources

When adding a new resource:

1. Create the resource file in `internal/provider/resource_*.go`
2. Implement the `Resource` interface:
   - `Metadata` - Set the resource type name
   - `Schema` - Define the resource schema
   - `Create` - Implement resource creation
   - `Read` - Implement resource reading
   - `Update` - Implement resource updating
   - `Delete` - Implement resource deletion
   - `ImportState` - Implement resource import (optional but recommended)
3. Register the resource in `internal/provider/provider.go`
4. Add examples in `examples/resources/`
5. Add tests in `internal/provider/resource_*_test.go`
6. Update documentation

## Adding New Data Sources

When adding a new data source:

1. Create the data source file in `internal/provider/data_source_*.go`
2. Implement the `DataSource` interface:
   - `Metadata` - Set the data source type name
   - `Schema` - Define the data source schema
   - `Read` - Implement data reading
3. Register the data source in `internal/provider/provider.go`
4. Add examples in `examples/data-sources/`
5. Add tests in `internal/provider/data_source_*_test.go`
6. Update documentation

## Adding New API Methods

When adding support for new ISP Config API methods:

1. Add the corresponding structs to `internal/client/models.go`
2. Implement the API methods in `internal/client/client.go`
3. Follow the existing naming convention (e.g., `Add*`, `Get*`, `Update*`, `Delete*`)
4. Handle errors appropriately
5. Add tests

## Documentation

- Keep the README.md up to date
- Add examples for new features
- Document breaking changes in CHANGELOG.md
- Use clear, concise language

## Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Ensure all tests pass
- Update documentation as needed
- Keep commits atomic and well-described
- One feature/fix per pull request

## Reporting Bugs

When reporting bugs, please include:

- Provider version
- Terraform version
- ISP Config version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant configuration snippets
- Error messages/logs

## Feature Requests

Feature requests are welcome! Please:

- Clearly describe the feature
- Explain the use case
- Provide examples if possible

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Assume good intentions

## Questions?

If you have questions, feel free to:

- Open an issue for discussion
- Join community discussions
- Reach out to maintainers

Thank you for contributing!


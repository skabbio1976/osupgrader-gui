# Contributing to OSUpgrader

First off, thank you for considering contributing to OSUpgrader! This tool was created to help IT operations teams upgrade 1000+ Windows Servers efficiently, and contributions from the community help make it better for everyone.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps to reproduce the problem**
* **Provide specific examples** (anonymized server names, configurations, etc.)
* **Describe the behavior you observed and what you expected**
* **Include logs** from `debuglogg.txt` (run with `-d` flag) - **Remember to redact sensitive information!**
* **Environment details**: OS, Go version, vCenter version

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

* **Use a clear and descriptive title**
* **Provide a detailed description** of the suggested enhancement
* **Explain why this enhancement would be useful** to most OSUpgrader users
* **List some examples** of how it would be used

### Pull Requests

* Fill in the required template
* Follow the Go coding style (run `gofmt` before committing)
* Include appropriate test coverage for new features
* Update documentation (README.md, code comments) as needed
* Keep commits focused - one feature/fix per PR when possible

## Development Process

### Setting Up Your Development Environment

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/osupgrader-gui.git`
3. Add upstream remote: `git remote add upstream https://github.com/ORIGINAL-OWNER/osupgrader-gui.git`
4. Install dependencies: `go mod tidy`
5. Create a branch: `git checkout -b feature/my-awesome-feature`

### Code Style

* **Go**: Follow standard Go conventions
  - Run `gofmt` to format code
  - Run `go vet` to catch common errors
  - Use meaningful variable and function names
  - Add comments for exported functions and complex logic
  - All comments should be in English

* **PowerShell**:
  - Use PascalCase for functions
  - Use clear variable names with `$` prefix
  - Add comments for complex operations
  - All comments should be in English

### Testing

* Test your changes locally with different scenarios:
  - Both Server Core and Desktop Experience
  - Different vCenter versions
  - Multi-domain environments
  - Edge cases (low disk space, network issues, etc.)

* Use debug mode (`-d` flag) to verify detailed logging works correctly

### Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters
* Reference issues and pull requests after the first line

Example:
```
Add support for Windows Server 2025

- Detect Windows Server 2025 in OS validation
- Update target OS list in upgrade logic
- Add 2025-specific GVLK keys

Fixes #123
```

### Documentation

* Keep README.md up to date with new features
* Update code comments when changing functionality
* Add examples for new configuration options
* Document breaking changes clearly

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Community Guidelines

* Be respectful and inclusive
* Assume good intentions
* Provide constructive feedback
* Help newcomers get started
* Remember: we're all here to help IT teams do their job better!

## Security

If you discover a security vulnerability, please **DO NOT** open a public issue. Instead:

1. Email the maintainers privately
2. Provide detailed information about the vulnerability
3. Allow time for a fix before public disclosure

## Recognition

Contributors will be recognized in:
* The project README
* Release notes for significant contributions
* Git history (make sure your commits use your real name and email)

## Questions?

Feel free to open an issue with the label `question` if you have any questions about contributing!

---

**Thank you for making OSUpgrader better for everyone!** 🎉

Every contribution, no matter how small, helps IT operations teams around the world upgrade their infrastructure more efficiently.

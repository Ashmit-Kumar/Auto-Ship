# Contributing to Auto-Ship

Thanks for your interest in contributing to Auto-Ship. This guide explains how to report bugs, suggest features, and contribute code.

Getting started

- Fork the repository and clone your fork:
  git clone git@github.com:<your-username>/Auto-Ship.git
- Create a feature branch:
  git checkout -b feat/your-feature
- Install prerequisites (see README and docs):
  - Go 1.20+
  - Python 3.11+
  - Docker
  - MongoDB or a running Atlas instance

Coding standards

- Follow existing code style. Use gofmt/golines for Go and black/isort for Python.
- Add unit tests where possible and ensure they pass.

Workflow

- Open a Pull Request with a clear title and description of the change.
- Keep changes focused and add tests for logic when possible.
- PRs will be reviewed and merged after approval.

Reporting bugs

- Open an issue explaining steps to reproduce, expected behavior, and logs if available.

Feature requests

- Create an issue and tag it with `enhancement`.

Security

- If you find a security vulnerability, please do not open a public issue. Contact the maintainers privately.

Thank you for contributing!

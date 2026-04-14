# Security Policy

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, email the maintainer directly: joedaviesio@protonmail.com.

Include:

- A description of the issue
- Steps to reproduce
- Affected version(s)
- Any potential impact you've identified

You can expect an initial response within 72 hours. We'll coordinate disclosure once a fix is ready.

## Scope

In scope:

- The `tirith` CLI and proxy (`cmd/tirith`, `internal/`)
- The Python SDK (`sdks/python/`)
- The TypeScript SDK (`sdks/typescript/`)

Out of scope:

- Vulnerabilities in upstream providers (Anthropic, OpenAI, Google) — report to them directly.
- Issues requiring physical access to a user's machine.

## Supported Versions

Only the latest minor release receives security updates. Please upgrade before reporting.

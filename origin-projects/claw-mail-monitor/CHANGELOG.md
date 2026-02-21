# Changelog

All notable changes to this project will be documented in this file.
This project follows the principles of Keep a Changelog and Semantic Versioning.

## [Unreleased]

- Switch to HTTP API + CLI (Gin + Cobra); remove MCP.
- Add HTTP endpoints and CLI commands for accounts, logs, send, test, latest.
- Add IMAP polling configuration and enhanced logging.
- Add service management via kardianos/service and system install flow.
- Add release automation and install script for prebuilt binaries.
- Remove logs command and endpoint; status now reports config/log paths.
- Add webhook session_key config and version command.
- Aggregate latest emails across enabled accounts when no email is specified.
- Add latest --since filter for recent emails.
- Add webhook custom_payload override.

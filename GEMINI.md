# Project Environment Standards

This document defines the required environment and toolchain specifications for this project.

## Core Toolchain

All developers and automated environments must use the following tool versions as verified in the sandbox:

| Tool | Required Version | Purpose |
| :--- | :--- | :--- |
| **Python** | 3.12.10 | Primary runtime and logic execution. |
| **uv** | 0.7.2+ | Python package and project management. |
| **curl** | 8.12.1+ | Networking and API interactions. |
| **cat** | 8.32+ | File manipulation (GNU coreutils). |

## Environment Configuration

- **OS**: Windows (win32)
- **Sandbox**: The project utilizes a local sandbox environment located in `./sandbox`.
- **Git/MINGW64**: MINGW64 must be available at `sandbox/git/mingw64` for compatible shell utilities.
- **Python VENV**: The primary virtual environment is located at `./.venv`.

## Development Workflows

- Always verify the toolchain using the `tool-status-checker` skill before starting significant work.
- Adhere to the paths defined in `TOOLS.md` for binary execution.

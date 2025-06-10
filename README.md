# StructDiff

StructDiff is a cross-platform CLI tool for comparing structured data files (JSON, YAML, TOML, XML, INI, CSV, HCL) and showing differences in a human-readable format.

## Features

- Compare multiple file formats: JSON, YAML, TOML, XML, INI, CSV, HCL
- Remote file comparison over HTTPS
- Authentication support (Basic Auth and Bearer Token)
- File size limits and timeout configuration
- Colorized output
- JSON output format
- Summary of differences
- Case-sensitive/insensitive comparison

## Installation

```bash
go install github.com/dolastack/structdiff@latest
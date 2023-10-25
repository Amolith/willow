# SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
#
# SPDX-License-Identifier: CC0-1.0

default: fmt lint staticcheck test vuln reuse

fmt:
    # Formatting all Go source code
    go install mvdan.cc/gofumpt@latest
    gofumpt -l -w .

lint:
    # Linting Go source code
    golangci-lint run

staticcheck:
    # Performing static analysis
    go install honnef.co/go/tools/cmd/staticcheck@latest
    staticcheck ./...

test:
    # Running tests
    go test -v ./...

vuln:
    # Checking for vulnerabilities
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...

reuse:
    # Linting licenses and copyright headers
    reuse lint

clean:
    # Cleaning up
    rm -rf willow out/

clean-all:
    # Removing build artifacts, willow.sqlite, config.toml, and data/ directory

    rm -rf willow out willow.sqlite config.toml data

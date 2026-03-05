# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Build**: `make build` - Build the gitea-mcp binary
**Install**: `make install` - Build and install to GOPATH/bin
**Clean**: `make clean` - Remove build artifacts
**Test**: `go test ./...` - Run all tests
**Hot reload**: `make dev` - Start development server with hot reload (requires air)
**Dependencies**: `make vendor` - Tidy and verify module dependencies

## Architecture Overview

This is a **Gitea MCP (Model Context Protocol) Server** written in Go that provides MCP tools for interacting with Gitea repositories, issues, pull requests, users, and more.

**Core Components**:

- `main.go` + `cmd/cmd.go`: CLI entry point and flag parsing
- `operation/operation.go`: Main server setup and tool registration
- `pkg/tool/tool.go`: Tool registry with read/write categorization
- `operation/*/`: Individual tool modules (user, repo, issue, pull, search, wiki, etc.)

**Transport Modes**:

- **stdio** (default): Standard input/output for MCP clients
- **http**: HTTP server mode on configurable port (default 8080)

**Authentication**:

- Global token via `--token` flag or `GITEA_ACCESS_TOKEN` env var
- HTTP mode supports per-request Bearer token override in Authorization header
- Token precedence: HTTP Authorization header > CLI flag > environment variable

**Tool Organization**:

- Tools are categorized as read-only or write operations
- `--read-only` flag exposes only read tools
- Tool modules register via `Tool.RegisterRead()` and `Tool.RegisterWrite()`

**Key Configuration**:

- Default Gitea host: `https://gitea.com` (override with `--host` or `GITEA_HOST`)
- Environment variables can override CLI flags: `MCP_MODE`, `GITEA_READONLY`, `GITEA_DEBUG`, `GITEA_INSECURE`
- Logs are written to `~/.gitea-mcp/gitea-mcp.log` with rotation

## Available Tools

The server provides 40+ MCP tools covering:

- **User**: get_my_user_info, get_user_orgs, search_users
- **Repository**: create_repo, fork_repo, list_my_repos, search_repos
- **Branches/Tags**: create_branch, delete_branch, list_branches, create_tag, list_tags
- **Files**: get_file_content, create_file, update_file, delete_file, get_dir_content
- **Issues**: create_issue, list_repo_issues, create_issue_comment, edit_issue
- **Pull Requests**: create_pull_request, list_repo_pull_requests, get_pull_request_by_index
- **Releases**: create_release, list_releases, get_latest_release
- **Wiki**: create_wiki_page, update_wiki_page, list_wiki_pages
- **Search**: search_repos, search_users, search_org_teams
- **Version**: get_gitea_mcp_server_version

## Common Development Patterns

**Testing**: Use `go test ./operation -run TestFunctionName` for specific tests

**Token Context**: HTTP requests use `pkg/context.TokenContextKey` for request-scoped token access

**Flag Access**: All packages access configuration via global variables in `pkg/flag/flag.go`

**Graceful Shutdown**: HTTP mode implements graceful shutdown with 10-second timeout on SIGTERM/SIGINT

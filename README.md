# Gitea MCP Compact

This is an **optimized fork** of the official **Gitea MCP Server** that adds compact format for commands like list and search.
**Large MCP responses can fill token context quickly**, this fork reduces token usage by **70-80%**.
It has other improvements, like ability to manage dependencies for issues.

Changes compared to original version:

- Original commands now act as compact version, but you can still use \_full variants
- Message text - even for \_full version - still truncated to 300 characters by default
- You can change amount of characters you want to truncate for both versions in settings
  - settings.json now used to store settings, but old MCP flags for host and token are still available
- agent_token in settings used as main way for agent to access Gitea (with agent username)
- You can add user_token for your own user in settings as well
  - In that case list_my_repos command going to return both your and agent's repos together, same for list_my_repos_full (including repos from organizations)
  - You can also set **"create_repo_as_user": true** to make agent create repos using your username instead
- New commands: **list_issue_dependencies, add_issue_dependency, remove_issue_dependency**

⚠️ Security Warning: Gitea doesn't have granular access scopes, so your agent can do anything, including connecting to API and deleting repositories, its up to you to resolve this issue using backups, separate Gitea server and restrictive prompts

Original version: https://gitea.com/gitea/gitea-mcp

**Gitea MCP Server** is an integration plugin designed to connect Gitea with Model Context Protocol (MCP) systems. This allows for seamless command execution and repository management through an MCP-compatible chat interface.

## Table of Contents

- [Gitea MCP Server](#gitea-mcp-server)
  - [Table of Contents](#table-of-contents)
  - [What is Gitea?](#what-is-gitea)
  - [What is MCP?](#what-is-mcp)
  - [🚧 Installation](#-installation)
    - [Usage with Claude Code](#usage-with-claude-code)
    - [Usage with VS Code](#usage-with-vs-code)
    - [📥 Download the official binary release](#-download-the-official-binary-release)
    - [🔧 Build from Source](#-build-from-source)
    - [📁 Add to PATH](#-add-to-path)
  - [🚀 Usage](#-usage)
  - [✅ Available Tools](#-available-tools)
  - [🐛 Debugging](#-debugging)
  - [🛠 Troubleshooting](#-troubleshooting)

## What is Gitea?

Gitea is a community-managed lightweight code hosting solution written in Go. It is published under the MIT license. Gitea provides Git hosting including a repository viewer, issue tracking, pull requests, and more.

## What is MCP?

Model Context Protocol (MCP) is a protocol that allows for the integration of various tools and systems through a chat interface. It enables seamless command execution and management of repositories, users, and other resources.

## 🚧 Installation

### Usage with Claude Code

This method uses `go run` and requires [Go](https://go.dev) to be installed.

**Recommended approach** (MCP server manages its own credentials):

1. Create and login as new Agent user in Gitea, then generate token in Settings > Application with full Read/Write access

2. Create settings file at `~/.gitea-mcp/settings.json` (or `C:\Users\<username>\.gitea-mcp\settings.json` on Windows), using your agent token - user_token is optional:

```json
{
  "agent_token": "your-gitea-token-here",
  "user_token": "",
  "host": "https://localhost:3000",
  "truncate_compact": 100,
  "truncate_full": 300,
  "create_repo_as_user": false
}
```

> **⚠️ Security Warning**: Gitea doesn't have granular access scopes for tokens. Your agent can do anything with full access, including connecting to the API and deleting repositories. It's up to you to mitigate this risk using:
> - Regular backups
> - Separate Gitea server for agent operations
> - Restrictive prompts that limit agent actions

3. Add MCP server to Claude Code:

- Backup and open file `~\.claude.json` (Windows) or `~/.config/claude/claude.json` (Mac/Linux)
- Add this code at the end of the file (before closing bracket) and insert your username:

```json
  "mcpServers": {
    "gitea": {
      "command": "C:\\Users\\<username>\\.gitea-mcp\\gitea-mcp.exe",
      "args": ["-t", "stdio"]
    }
  }
```

- Put gitea-mcp.exe in this folder, then restart Claude Code and ask agent to use mcp commands

### 🔧 Build from Source

You can download the source code by cloning the repository using Git:

```bash
git clone https://github.com/appslop/gitea-mcp-compact.git
```

Before building, make sure you have the following installed:

- make
- Golang (Go 1.24 or later recommended)

Then run (Mac/Linux):

```bash
make install
```

Windows:

```bash
.\build.ps1 build
```

### 📁 Add to PATH

After installing, copy the binary gitea-mcp to a directory included in your system's PATH. For example:

```bash
cp gitea-mcp /usr/local/bin/
```

**Default log path**: `$HOME/.gitea-mcp/gitea-mcp.log`

> [!NOTE]
> You can provide your Gitea host and access token via HTTP headers (highest priority), command-line arguments, settings file, or environment variables.
> HTTP headers have the highest priority, followed by command-line arguments

Once everything is set up, try typing the following in your MCP-compatible chatbox:

```text
list all my repositories
```

**CLI flags:**

```bash
gitea-mcp -T agent_token_here --user-token user_token_here
```

**Environment variables:**

```bash
export GITEA_ACCESS_TOKEN=agent_token_here
export GITEA_USER_ACCESS_TOKEN=user_token_here
```

**HTTP mode with headers:**

```json
{
  "mcpServers": {
    "gitea": {
      "url": "http://localhost:3000/mcp",
      "headers": {
        "Authorization": "Bearer agent_token_here",
        "X-User-Token": "user_token_here"
      }
    }
  }
}
```

**Token Precedence** (for both agent and user tokens):

1. HTTP header (highest priority)
2. CLI flag
3. Settings file (`~/.gitea-mcp/settings.json`)
4. Environment variable (lowest priority)

### When to use each token?

- **agent_token**: Default for all operations
- **user_token**: Use with `create_repo_as_user=true` to create repos as your user instead of agent
- **list_my_repos_full**: Automatically combines repos from both tokens

## ✅ Available Tools

The Gitea MCP Server supports the following tools:

### Compact vs Full Format

List and search tools come in two variants to optimize token usage:

- **Compact format** (default): Returns essential information for browsing and filtering, reducing token usage by 70-80%
  - Long text fields truncated to 100 characters by default (configurable)
  - Nested objects converted to simple strings
  - Metadata and URLs removed
  - **Issues/PRs default to "open" state** (active items, most commonly needed)

- **Full format** (`*_full` variant): Returns complete data when you need detailed information
  - Contains all fields, URLs, and metadata
  - **Issues/PRs default to "all" states** (open + closed, for comprehensive analysis)
  - Still truncated to 300 chars by default, if you need full message text - set truncate_full to very high value (10000) in settings

**Example workflow**: Search repos using compact format to find interesting ones → use `get_file_content` or specific get-by-ID tools for details → only use full format when you need comprehensive information about a specific item.

|               Tool                |    Scope     |                           Description                           |
| :-------------------------------: | :----------: | :-------------------------------------------------------------: |
|         get_my_user_info          |     User     |          Get the information of the authenticated user          |
|           get_user_orgs           |     User     |    Get organizations associated with the authenticated user     |
|            create_repo            |  Repository  |                     Create a new repository                     |
|             fork_repo             |  Repository  |                        Fork a repository                        |
|           list_my_repos           |  Repository  | List all repositories owned by the authenticated user (compact) |
|        list_my_repos_full         |  Repository  |  List all repositories owned by the authenticated user (full)   |
|           create_branch           |    Branch    |                       Create a new branch                       |
|           delete_branch           |    Branch    |                         Delete a branch                         |
|           list_branches           |    Branch    |           List all branches in a repository (compact)           |
|        list_branches_full         |    Branch    |            List all branches in a repository (full)             |
|          create_release           |   Release    |              Create a new release in a repository               |
|          delete_release           |   Release    |               Delete a release from a repository                |
|            get_release            |   Release    |                          Get a release                          |
|        get_latest_release         |   Release    |             Get the latest release in a repository              |
|           list_releases           |   Release    |           List all releases in a repository (compact)           |
|        list_releases_full         |   Release    |            List all releases in a repository (full)             |
|            create_tag             |     Tag      |                        Create a new tag                         |
|            delete_tag             |     Tag      |                          Delete a tag                           |
|              get_tag              |     Tag      |                            Get a tag                            |
|             list_tags             |     Tag      |                  List all tags in a repository                  |
|         list_repo_commits         |    Commit    |           List all commits in a repository (compact)            |
|      list_repo_commits_full       |    Commit    |             List all commits in a repository (full)             |
|         get_file_content          |     File     |             Get the content and metadata of a file              |
|          get_dir_content          |     File     |              Get a list of entries in a directory               |
|            create_file            |     File     |                        Create a new file                        |
|            update_file            |     File     |                     Update an existing file                     |
|            delete_file            |     File     |                          Delete a file                          |
|        get_issue_by_index         |    Issue     |                    Get an issue by its index                    |
|         list_repo_issues          |    Issue     |      List open issues (state='all' to show all) (compact)       |
|       list_repo_issues_full       |    Issue     |             List all issues in a repository (full)              |
|           create_issue            |    Issue     |                       Create a new issue                        |
|       create_issue_comment        |    Issue     |                  Create a comment on an issue                   |
|            edit_issue             |    Issue     |                          Edit a issue                           |
|        edit_issue_comment         |    Issue     |                   Edit a comment on an issue                    |
|    get_issue_comments_by_index    |    Issue     |              Get comments of an issue by its index              |
|      list_issue_dependencies      |    Issue     |      List all dependencies (blocking issues) for an issue       |
|       add_issue_dependency        |    Issue     |                  Add a dependency to an issue                   |
|      remove_issue_dependency      |    Issue     |                Remove a dependency from an issue                |
|     get_pull_request_by_index     | Pull Request |                 Get a pull request by its index                 |
|       get_pull_request_diff       | Pull Request |                     Get a pull request diff                     |
|      list_repo_pull_requests      | Pull Request |   List open pull requests (state='all' to show all) (compact)   |
|   list_repo_pull_requests_full    | Pull Request |          List all pull requests in a repository (full)          |
|        create_pull_request        | Pull Request |                    Create a new pull request                    |
|   create_pull_request_reviewer    | Pull Request |                 Add reviewers to a pull request                 |
|   delete_pull_request_reviewer    | Pull Request |              Remove reviewers from a pull request               |
|     list_pull_request_reviews     | Pull Request |               List all reviews for a pull request               |
|      get_pull_request_review      | Pull Request |                   Get a specific review by ID                   |
| list_pull_request_review_comments | Pull Request |                List inline comments for a review                |
|    create_pull_request_review     | Pull Request |          Create a review with optional inline comments          |
|    submit_pull_request_review     | Pull Request |                     Submit a pending review                     |
|    delete_pull_request_review     | Pull Request |                         Delete a review                         |
|    dismiss_pull_request_review    | Pull Request |             Dismiss a review with optional message              |
|        merge_pull_request         | Pull Request |                      Merge a pull request                       |
|           search_users            |     User     |                   Search for users (compact)                    |
|         search_users_full         |     User     |                     Search for users (full)                     |
|         search_org_teams          | Organization |               Search for teams in an organization               |
|          list_org_labels          | Organization |            List labels defined at organization level            |
|         create_org_label          | Organization |                Create a label in an organization                |
|          edit_org_label           | Organization |                 Edit a label in an organization                 |
|         delete_org_label          | Organization |                Delete a label in an organization                |
|           search_repos            |  Repository  |                Search for repositories (compact)                |
|         search_repos_full         |  Repository  |                 Search for repositories (full)                  |
|     list_repo_action_secrets      |   Actions    |         List repository Actions secrets (metadata only)         |
|     upsert_repo_action_secret     |   Actions    |       Create/update (upsert) a repository Actions secret        |
|     delete_repo_action_secret     |   Actions    |               Delete a repository Actions secret                |
|      list_org_action_secrets      |   Actions    |        List organization Actions secrets (metadata only)        |
|     upsert_org_action_secret      |   Actions    |      Create/update (upsert) an organization Actions secret      |
|     delete_org_action_secret      |   Actions    |              Delete an organization Actions secret              |
|    list_repo_action_variables     |   Actions    |                List repository Actions variables                |
|     get_repo_action_variable      |   Actions    |                Get a repository Actions variable                |
|    create_repo_action_variable    |   Actions    |              Create a repository Actions variable               |
|    update_repo_action_variable    |   Actions    |              Update a repository Actions variable               |
|    delete_repo_action_variable    |   Actions    |              Delete a repository Actions variable               |
|     list_org_action_variables     |   Actions    |               List organization Actions variables               |
|      get_org_action_variable      |   Actions    |              Get an organization Actions variable               |
|    create_org_action_variable     |   Actions    |             Create an organization Actions variable             |
|    update_org_action_variable     |   Actions    |             Update an organization Actions variable             |
|    delete_org_action_variable     |   Actions    |             Delete an organization Actions variable             |
|    list_repo_action_workflows     |   Actions    |                List repository Actions workflows                |
|     get_repo_action_workflow      |   Actions    |                Get a repository Actions workflow                |
|   dispatch_repo_action_workflow   |   Actions    |        Trigger (dispatch) a repository Actions workflow         |
|       list_repo_action_runs       |   Actions    |                  List repository Actions runs                   |
|        get_repo_action_run        |   Actions    |                  Get a repository Actions run                   |
|      cancel_repo_action_run       |   Actions    |                 Cancel a repository Actions run                 |
|       rerun_repo_action_run       |   Actions    |                 Rerun a repository Actions run                  |
|       list_repo_action_jobs       |   Actions    |                  List repository Actions jobs                   |
|     list_repo_action_run_jobs     |   Actions    |                   List Actions jobs for a run                   |
|  get_repo_action_job_log_preview  |   Actions    |              Get a job log preview (tail/limited)               |
|   download_repo_action_job_log    |   Actions    |                  Download a job log to a file                   |
|   get_gitea_mcp_server_version    |    Server    |             Get the version of the Gitea MCP Server             |
|          list_wiki_pages          |     Wiki     |               List all wiki pages in a repository               |
|           get_wiki_page           |     Wiki     |              Get a wiki page content and metadata               |
|        get_wiki_revisions         |     Wiki     |              Get revisions history of a wiki page               |
|         create_wiki_page          |     Wiki     |                     Create a new wiki page                      |
|         update_wiki_page          |     Wiki     |                  Update an existing wiki page                   |
|         delete_wiki_page          |     Wiki     |                       Delete a wiki page                        |

## 🐛 Debugging

To enable debug mode, add the `-d` flag when running the Gitea MCP Server with http mode:

```sh
./gitea-mcp -t http [--port 3000] --token <your personal access token> -d
```

## 🛠 Troubleshooting

If you encounter any issues, here are some common troubleshooting steps:

1. **Check your PATH**: Ensure that the `gitea-mcp` binary is in a directory included in your system's PATH.
2. **Verify dependencies**: Make sure you have all the required dependencies installed, such as `make` and `Golang`.
3. **Review configuration**: Double-check your MCP configuration file for any errors or missing information.
4. **Consult logs**: Check the logs for any error messages or warnings that can provide more information about the issue.

Enjoy exploring and managing your Gitea repositories via chat!

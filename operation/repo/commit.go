package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	flagPkg "gitea.com/gitea/gitea-mcp/pkg/flag"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ListRepoCommitsToolName    = "list_repo_commits"
	ListRepoCommitsFullToolName = "list_repo_commits_full"
)

// CompactCommit represents a minimal commit object for efficient listing
type CompactCommit struct {
	SHA            string `json:"sha"`
	Message        string `json:"message,omitempty"`
	AuthorName     string `json:"author_name,omitempty"`
	AuthorEmail    string `json:"author_email,omitempty"`
	CommitterName  string `json:"committer_name,omitempty"`
	CommitterEmail string `json:"committer_email,omitempty"`
	Created        string `json:"created,omitempty"`
}

// toCompactCommits converts full SDK commits to compact commits
func toCompactCommits(commits []*gitea_sdk.Commit) []CompactCommit {
	result := make([]CompactCommit, len(commits))
	for i, commit := range commits {
		sha := commit.SHA

		// Get first line of message from RepoCommit
		message := ""
		if commit.RepoCommit != nil {
			message = commit.RepoCommit.Message
			if idx := strings.IndexByte(message, '\n'); idx > 0 {
				message = message[:idx]
			}
		}

		// Get author info from RepoCommit
		authorName := ""
		authorEmail := ""
		if commit.RepoCommit != nil && commit.RepoCommit.Author != nil {
			authorName = commit.RepoCommit.Author.Name
			authorEmail = commit.RepoCommit.Author.Email
		}

		// Get committer info from RepoCommit
		committerName := ""
		committerEmail := ""
		if commit.RepoCommit != nil && commit.RepoCommit.Committer != nil {
			committerName = commit.RepoCommit.Committer.Name
			committerEmail = commit.RepoCommit.Committer.Email
		}

		// Format created date
		created := ""
		if !commit.Created.IsZero() {
			created = commit.Created.Format("2006-01-02 15:04:05")
		}

		result[i] = CompactCommit{
			SHA:            sha,
			Message:        message,
			AuthorName:     authorName,
			AuthorEmail:    authorEmail,
			CommitterName:  committerName,
			CommitterEmail: committerEmail,
			Created:        created,
		}
	}
	return result
}

// truncateCommitMessage truncates the message field of commits for full format
func truncateCommitMessage(commits []*gitea_sdk.Commit) {
	for _, commit := range commits {
		if commit.RepoCommit != nil && commit.RepoCommit.Message != "" {
			if len(commit.RepoCommit.Message) > flagPkg.TruncateFull {
				commit.RepoCommit.Message = commit.RepoCommit.Message[:flagPkg.TruncateFull] + "..."
			}
		}
	}
}


var ListRepoCommitsTool = mcp.NewTool(
	ListRepoCommitsToolName,
	mcp.WithDescription("List repository commits (compact format, optimized for token usage)"),
	mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
	mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	mcp.WithString("sha", mcp.Description("SHA or branch to start listing commits from")),
	mcp.WithString("path", mcp.Description("path indicates that only commits that include the path's file/dir should be returned.")),
	mcp.WithNumber("page", mcp.Required(), mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
	mcp.WithNumber("page_size", mcp.Required(), mcp.Description("page size"), mcp.DefaultNumber(50), mcp.Min(1)),
)

var ListRepoCommitsFullTool = mcp.NewTool(
	ListRepoCommitsFullToolName,
	mcp.WithDescription("List repository commits (full format with all metadata, use when detailed info needed)"),
	mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
	mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	mcp.WithString("sha", mcp.Description("SHA or branch to start listing commits from")),
	mcp.WithString("path", mcp.Description("path indicates that only commits that include the path's file/dir should be returned.")),
	mcp.WithNumber("page", mcp.Required(), mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
	mcp.WithNumber("page_size", mcp.Required(), mcp.Description("page size"), mcp.DefaultNumber(50), mcp.Min(1)),
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListRepoCommitsTool,
		Handler: ListRepoCommitsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListRepoCommitsFullTool,
		Handler: ListRepoCommitsFullFn,
	})
}

func ListRepoCommitsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoCommitsFn (compact)")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	page, err := params.GetIndex(req.GetArguments(), "page")
	if err != nil {
		return to.ErrorResult(err)
	}
	pageSize, err := params.GetIndex(req.GetArguments(), "page_size")
	if err != nil {
		return to.ErrorResult(err)
	}
	sha, _ := req.GetArguments()["sha"].(string)
	path, _ := req.GetArguments()["path"].(string)
	opt := gitea_sdk.ListCommitOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
		SHA:  sha,
		Path: path,
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	commits, _, err := client.ListRepoCommits(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repo commits err: %v", err))
	}

	// Convert to compact format
	compactCommits := toCompactCommits(commits)
	return to.TextResult(compactCommits)
}

func ListRepoCommitsFullFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoCommitsFullFn (full)")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	page, err := params.GetIndex(req.GetArguments(), "page")
	if err != nil {
		return to.ErrorResult(err)
	}
	pageSize, err := params.GetIndex(req.GetArguments(), "page_size")
	if err != nil {
		return to.ErrorResult(err)
	}
	sha, _ := req.GetArguments()["sha"].(string)
	path, _ := req.GetArguments()["path"].(string)
	opt := gitea_sdk.ListCommitOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
		SHA:  sha,
		Path: path,
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	commits, _, err := client.ListRepoCommits(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repo commits err: %v", err))
	}

	// Truncate message for full format
	truncateCommitMessage(commits)

	// Return full format
	return to.TextResult(commits)
}

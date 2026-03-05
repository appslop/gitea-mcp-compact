package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	CreateBranchToolName    = "create_branch"
	DeleteBranchToolName    = "delete_branch"
	ListBranchesToolName    = "list_branches"
	ListBranchesFullToolName = "list_branches_full"
)

// CompactBranch represents a minimal branch object for efficient listing
type CompactBranch struct {
	Name           string `json:"name"`
	CommitSHA      string `json:"commit_sha"`
	CommitMessage  string `json:"commit_message,omitempty"`
	Author         string `json:"author,omitempty"`
	CommitterEmail string `json:"committer_email,omitempty"`
	Created        string `json:"created,omitempty"`
}

// toCompactBranches converts full SDK branches to compact branches
func toCompactBranches(branches []*gitea_sdk.Branch) []CompactBranch {
	result := make([]CompactBranch, len(branches))
	for i, branch := range branches {
		// Use full SHA from commit
		sha := ""
		if branch.Commit != nil {
			sha = branch.Commit.ID
		}
		// Get first line of commit message
		message := ""
		author := ""
		created := ""
		if branch.Commit != nil {
			message = branch.Commit.Message
			// Extract first line only
			if idx := strings.IndexByte(message, '\n'); idx > 0 {
				message = message[:idx]
			}

			// Get author name
			if branch.Commit.Author != nil {
				author = branch.Commit.Author.Name
			}

			// Get created date from timestamp
			if !branch.Commit.Timestamp.IsZero() {
				created = branch.Commit.Timestamp.Format("2006-01-02 15:04:05")
			}
		}

		result[i] = CompactBranch{
			Name:           branch.Name,
			CommitSHA:      sha,
			CommitMessage:  message,
			Author:         author,
			CommitterEmail: "",
			Created:        created,
		}
	}
	return result
}

var (
	CreateBranchTool = mcp.NewTool(
		CreateBranchToolName,
		mcp.WithDescription("Create branch"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Name of the branch to create")),
		mcp.WithString("old_branch", mcp.Required(), mcp.Description("Name of the old branch to create from")),
	)

	DeleteBranchTool = mcp.NewTool(
		DeleteBranchToolName,
		mcp.WithDescription("Delete branch"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Name of the branch to delete")),
	)

	ListBranchesTool = mcp.NewTool(
		ListBranchesToolName,
		mcp.WithDescription("List branches (compact format, optimized for token usage)"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)

	ListBranchesFullTool = mcp.NewTool(
		ListBranchesFullToolName,
		mcp.WithDescription("List branches (full format with all metadata, use when detailed info needed)"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)
)

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateBranchTool,
		Handler: CreateBranchFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteBranchTool,
		Handler: DeleteBranchFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListBranchesTool,
		Handler: ListBranchesFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListBranchesFullTool,
		Handler: ListBranchesFullFn,
	})
}

func CreateBranchFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateBranchFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	branch, ok := req.GetArguments()["branch"].(string)
	if !ok {
		return to.ErrorResult(errors.New("branch is required"))
	}
	oldBranch, _ := req.GetArguments()["old_branch"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, _, err = client.CreateBranch(owner, repo, gitea_sdk.CreateBranchOption{
		BranchName:    branch,
		OldBranchName: oldBranch,
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create branch error: %v", err))
	}

	return mcp.NewToolResultText("Branch Created"), nil
}

func DeleteBranchFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteBranchFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	branch, ok := req.GetArguments()["branch"].(string)
	if !ok {
		return to.ErrorResult(errors.New("branch is required"))
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, _, err = client.DeleteRepoBranch(owner, repo, branch)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete branch error: %v", err))
	}

	return to.TextResult("Branch Deleted")
}

func ListBranchesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListBranchesFn (compact)")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	opt := gitea_sdk.ListRepoBranchesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	branches, _, err := client.ListRepoBranches(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list branches error: %v", err))
	}

	// Convert to compact format
	compactBranches := toCompactBranches(branches)
	return to.TextResult(compactBranches)
}

func ListBranchesFullFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListBranchesFullFn (full)")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	opt := gitea_sdk.ListRepoBranchesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	branches, _, err := client.ListRepoBranches(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list branches error: %v", err))
	}

	// Return full format
	return to.TextResult(branches)
}

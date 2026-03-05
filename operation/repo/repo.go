package repo

import (
	"context"
	"errors"
	"fmt"

	flagPkg "gitea.com/gitea/gitea-mcp/pkg/flag"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"
	"gitea.com/gitea/gitea-mcp/pkg/tool"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var Tool = tool.New()

const (
	CreateRepoToolName     = "create_repo"
	ForkRepoToolName       = "fork_repo"
	ListMyReposToolName    = "list_my_repos"
	ListMyReposFullToolName = "list_my_repos_full"
)

// CompactRepository represents a minimal repository object for efficient listing
type CompactRepository struct {
	Name         string `json:"name"`
	Owner        string `json:"owner,omitempty"`
	FullName     string `json:"full_name"`
	Description  string `json:"description,omitempty"`
	Private      bool   `json:"private"`
	StarCount    int    `json:"stars_count"`
	ForksCount   int    `json:"forks_count"`
	OpenIssuesCount int  `json:"open_issues_count"`
	Language     string `json:"language,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	Created      string `json:"created_at,omitempty"`
	Archived     bool   `json:"archived"`
}

// toCompactRepos converts full SDK repositories to compact repositories
func toCompactRepos(repos []*gitea_sdk.Repository) []CompactRepository {
	result := make([]CompactRepository, len(repos))
	for i, repo := range repos {
		// Truncate description to preview (first 200 chars)
		descPreview := ""
		if repo.Description != "" {
			if len(repo.Description) > 200 {
				descPreview = repo.Description[:200] + "..."
			} else {
				descPreview = repo.Description
			}
		}

		// Get owner name
		owner := ""
		if repo.Owner != nil {
			owner = repo.Owner.UserName
		}

		result[i] = CompactRepository{
			Name:            repo.Name,
			Owner:           owner,
			FullName:        repo.FullName,
			Description:     descPreview,
			Private:         repo.Private,
			StarCount:       repo.Stars,
			ForksCount:      repo.Forks,
			OpenIssuesCount: repo.OpenIssues,
			Language:        repo.Language,
			UpdatedAt:       repo.Updated.Format("2006-01-02 15:04:05"),
			Created:         repo.Created.Format("2006-01-02 15:04:05"),
			Archived:        repo.Archived,
		}
	}
	return result
}

var (
	CreateRepoTool = mcp.NewTool(
		CreateRepoToolName,
		mcp.WithDescription("Create repository in personal account or organization"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the repository to create")),
		mcp.WithString("description", mcp.Description("Description of the repository to create")),
		mcp.WithBoolean("private", mcp.Description("Whether the repository is private")),
		mcp.WithString("issue_labels", mcp.Description("Issue Label set to use")),
		mcp.WithBoolean("auto_init", mcp.Description("Whether the repository should be auto-intialized?")),
		mcp.WithBoolean("template", mcp.Description("Whether the repository is template")),
		mcp.WithString("gitignores", mcp.Description("Gitignores to use")),
		mcp.WithString("license", mcp.Description("License to use")),
		mcp.WithString("readme", mcp.Description("Readme template name (e.g., 'Default'). Leave empty for no readme, or use create_file/update_file after repo creation to add custom content")),
		mcp.WithString("default_branch", mcp.Description("DefaultBranch of the repository (used when initializes and in template)")),
		mcp.WithString("organization", mcp.Description("Organization name to create repository in (optional - defaults to personal account)")),
	)

	ForkRepoTool = mcp.NewTool(
		ForkRepoToolName,
		mcp.WithDescription("Fork repository"),
		mcp.WithString("user", mcp.Required(), mcp.Description("User name of the repository to fork")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name to fork")),
		mcp.WithString("organization", mcp.Description("Organization name to fork")),
		mcp.WithString("name", mcp.Description("Name of the forked repository")),
	)

	ListMyReposTool = mcp.NewTool(
		ListMyReposToolName,
		mcp.WithDescription("List my repositories (compact format, optimized for token usage)"),
		mcp.WithNumber("page", mcp.Description("Page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("pageSize", mcp.Description("Page size"), mcp.DefaultNumber(100), mcp.Min(1)),
	)

	ListMyReposFullTool = mcp.NewTool(
		ListMyReposFullToolName,
		mcp.WithDescription("List my repositories (full format with all metadata, use when detailed info needed)"),
		mcp.WithNumber("page", mcp.Description("Page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("pageSize", mcp.Description("Page size"), mcp.DefaultNumber(100), mcp.Min(1)),
	)
)

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateRepoTool,
		Handler: CreateRepoFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    ForkRepoTool,
		Handler: ForkRepoFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListMyReposTool,
		Handler: ListMyReposFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListMyReposFullTool,
		Handler: ListMyReposFullFn,
	})
}

func CreateRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateRepoFn")
	name, ok := req.GetArguments()["name"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repository name is required"))
	}
	description, _ := req.GetArguments()["description"].(string)
	private, _ := req.GetArguments()["private"].(bool)
	issueLabels, _ := req.GetArguments()["issue_labels"].(string)
	autoInit, _ := req.GetArguments()["auto_init"].(bool)
	template, _ := req.GetArguments()["template"].(bool)
	gitignores, _ := req.GetArguments()["gitignores"].(string)
	license, _ := req.GetArguments()["license"].(string)
	readme, _ := req.GetArguments()["readme"].(string)
	defaultBranch, _ := req.GetArguments()["default_branch"].(string)
	organization, _ := req.GetArguments()["organization"].(string)

	opt := gitea_sdk.CreateRepoOption{
		Name:          name,
		Description:   description,
		Private:       private,
		IssueLabels:   issueLabels,
		AutoInit:      autoInit,
		Template:      template,
		Gitignores:    gitignores,
		License:       license,
		Readme:        readme,
		DefaultBranch: defaultBranch,
	}

	var repo *gitea_sdk.Repository
	var client *gitea_sdk.Client
	var err error

	// Choose client based on CreateRepoAsUser flag
	if flagPkg.CreateRepoAsUser {
		log.Debugf("Creating repo as user (using user_token)")
		client, err = gitea.UserClientFromContext(ctx)
	} else {
		log.Debugf("Creating repo as agent (using token)")
		client, err = gitea.ClientFromContext(ctx)
	}
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	if organization != "" {
		repo, _, err = client.CreateOrgRepo(organization, opt)
		if err != nil {
			return to.ErrorResult(fmt.Errorf("create organization repository '%s' in '%s' err: %v", name, organization, err))
		}
	} else {
		repo, _, err = client.CreateRepo(opt)
		if err != nil {
			return to.ErrorResult(fmt.Errorf("create repository '%s' err: %v", name, err))
		}
	}
	return to.TextResult(repo)
}

func ForkRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ForkRepoFn")
	user, ok := req.GetArguments()["user"].(string)
	if !ok {
		return to.ErrorResult(errors.New("user name is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repository name is required"))
	}
	organization, _ := req.GetArguments()["organization"].(string)
	var organizationPtr *string
	if organization != "" {
		organizationPtr = &organization
	}
	name, _ := req.GetArguments()["name"].(string)
	var namePtr *string
	if name != "" {
		namePtr = &name
	}
	opt := gitea_sdk.CreateForkOption{
		Organization: organizationPtr,
		Name:         namePtr,
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, _, err = client.CreateFork(user, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("fork repository error: %v", err))
	}
	return to.TextResult("Fork success")
}

func ListMyReposFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListMyReposFn (compact)")
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.ListReposOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	repos, _, err := client.ListMyRepos(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list my repositories error: %v", err))
	}

	// Convert to compact format
	compactRepos := toCompactRepos(repos)
	return to.TextResult(compactRepos)
}

func ListMyReposFullFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListMyReposFullFn (full)")
	
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.ListReposOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}

	var allRepos []*gitea_sdk.Repository

	// Get repos from agent token (regular user)
	client1, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	log.Debugf("Fetching repos for agent token (including orgs)")
	agentRepos, _, err := client1.ListMyRepos(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repos for agent token error: %v", err))
	}
	allRepos = append(allRepos, agentRepos...)

	// Get repos from user token (if available)
	client2, err := gitea.UserClientFromContext(ctx)
	if err == nil && client2 != nil {
		log.Debugf("Fetching repos for user token (including orgs)")
		userRepos, _, err := client2.ListMyRepos(opt)
		if err != nil {
			log.Debugf("Could not list repos for user token: %v", err)
		} else {
			allRepos = append(allRepos, userRepos...)
		}
	} else {
		log.Debugf("No user token available, skipping user repos")
	}

	// Return full format
	return to.TextResult(allRepos)
}

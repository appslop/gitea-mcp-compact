package search

import (
	"context"
	"errors"
	"fmt"

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
	SearchUsersToolName      = "search_users"
	SearchUsersFullToolName  = "search_users_full"
	SearchOrgTeamsToolName   = "search_org_teams"
	SearchReposToolName      = "search_repos"
	SearchReposFullToolName  = "search_repos_full"
)

// CompactUser represents a minimal user object for efficient search results
type CompactUser struct {
	UserName  string `json:"username"`
	FullName  string `json:"full_name,omitempty"`
	Email     string `json:"email,omitempty"`
	IsActive  bool   `json:"is_active"`
	IsAdmin   bool   `json:"is_admin"`
	Created   string `json:"created,omitempty"`
	Location  string `json:"location,omitempty"`
	Website   string `json:"website,omitempty"`
}

// toCompactUsers converts full SDK users to compact users
func toCompactUsers(users []*gitea_sdk.User) []CompactUser {
	result := make([]CompactUser, len(users))
	for i, user := range users {
		// Format created date
		created := ""
		if !user.Created.IsZero() {
			created = user.Created.Format("2006-01-02 15:04:05")
		}

		result[i] = CompactUser{
			UserName: user.UserName,
			FullName: user.FullName,
			Email:    user.Email,
			IsActive: user.IsActive,
			IsAdmin:  user.IsAdmin,
			Created:  created,
			Location: user.Location,
			Website:  user.Website,
		}
	}
	return result
}

// CompactRepository represents a minimal repository object for efficient search results
type CompactRepository struct {
	Name            string `json:"name"`
	Owner           string `json:"owner,omitempty"`
	FullName        string `json:"full_name"`
	Description     string `json:"description,omitempty"`
	Private         bool   `json:"private"`
	StarCount       int    `json:"stars_count"`
	ForksCount      int    `json:"forks_count"`
	OpenIssuesCount int    `json:"open_issues_count"`
	Language        string `json:"language,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	Created         string `json:"created_at,omitempty"`
	Archived        bool   `json:"archived"`
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
	SearchUsersTool = mcp.NewTool(
		SearchUsersToolName,
		mcp.WithDescription("search users (compact format, optimized for token usage)"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("Keyword")),
		mcp.WithNumber("page", mcp.Description("Page"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("PageSize"), mcp.DefaultNumber(100)),
	)

	SearchUsersFullTool = mcp.NewTool(
		SearchUsersFullToolName,
		mcp.WithDescription("search users (full format with all metadata, use when detailed info needed)"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("Keyword")),
		mcp.WithNumber("page", mcp.Description("Page"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("PageSize"), mcp.DefaultNumber(100)),
	)

	SearOrgTeamsTool = mcp.NewTool(
		SearchOrgTeamsToolName,
		mcp.WithDescription("search organization teams"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithString("query", mcp.Required(), mcp.Description("search organization teams")),
		mcp.WithBoolean("includeDescription", mcp.Description("include description?")),
		mcp.WithNumber("page", mcp.Description("Page"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("PageSize"), mcp.DefaultNumber(100)),
	)

	SearchReposTool = mcp.NewTool(
		SearchReposToolName,
		mcp.WithDescription("search repos (compact format, optimized for token usage)"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("Keyword")),
		mcp.WithBoolean("keywordIsTopic", mcp.Description("KeywordIsTopic")),
		mcp.WithBoolean("keywordInDescription", mcp.Description("KeywordInDescription")),
		mcp.WithNumber("ownerID", mcp.Description("OwnerID")),
		mcp.WithBoolean("isPrivate", mcp.Description("IsPrivate")),
		mcp.WithBoolean("isArchived", mcp.Description("IsArchived")),
		mcp.WithString("sort", mcp.Description("Sort")),
		mcp.WithString("order", mcp.Description("Order")),
		mcp.WithNumber("page", mcp.Description("Page"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("PageSize"), mcp.DefaultNumber(100)),
	)

	SearchReposFullTool = mcp.NewTool(
		SearchReposFullToolName,
		mcp.WithDescription("search repos (full format with all metadata, use when detailed info needed)"),
		mcp.WithString("keyword", mcp.Required(), mcp.Description("Keyword")),
		mcp.WithBoolean("keywordIsTopic", mcp.Description("KeywordIsTopic")),
		mcp.WithBoolean("keywordInDescription", mcp.Description("KeywordInDescription")),
		mcp.WithNumber("ownerID", mcp.Description("OwnerID")),
		mcp.WithBoolean("isPrivate", mcp.Description("IsPrivate")),
		mcp.WithBoolean("isArchived", mcp.Description("IsArchived")),
		mcp.WithString("sort", mcp.Description("Sort")),
		mcp.WithString("order", mcp.Description("Order")),
		mcp.WithNumber("page", mcp.Description("Page"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("PageSize"), mcp.DefaultNumber(100)),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    SearchUsersTool,
		Handler: UsersFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    SearchUsersFullTool,
		Handler: UsersFullFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    SearOrgTeamsTool,
		Handler: OrgTeamsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    SearchReposTool,
		Handler: ReposFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    SearchReposFullTool,
		Handler: ReposFullFn,
	})
}

func UsersFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called UsersFn (compact)")
	keyword, ok := req.GetArguments()["keyword"].(string)
	if !ok {
		return to.ErrorResult(errors.New("keyword is required"))
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.SearchUsersOption{
		KeyWord: keyword,
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	users, _, err := client.SearchUsers(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("search users err: %v", err))
	}

	// Convert to compact format
	compactUsers := toCompactUsers(users)
	return to.TextResult(compactUsers)
}

func UsersFullFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called UsersFullFn (full)")
	keyword, ok := req.GetArguments()["keyword"].(string)
	if !ok {
		return to.ErrorResult(errors.New("keyword is required"))
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.SearchUsersOption{
		KeyWord: keyword,
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	users, _, err := client.SearchUsers(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("search users err: %v", err))
	}

	// Return full format
	return to.TextResult(users)
}

func OrgTeamsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called OrgTeamsFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok {
		return to.ErrorResult(errors.New("organization is required"))
	}
	query, ok := req.GetArguments()["query"].(string)
	if !ok {
		return to.ErrorResult(errors.New("query is required"))
	}
	includeDescription, _ := req.GetArguments()["includeDescription"].(bool)
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.SearchTeamsOptions{
		Query:              query,
		IncludeDescription: includeDescription,
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	teams, _, err := client.SearchOrgTeams(org, &opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("search organization teams error: %v", err))
	}
	return to.TextResult(teams)
}

func ReposFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ReposFn (compact)")
	keyword, ok := req.GetArguments()["keyword"].(string)
	if !ok {
		return to.ErrorResult(errors.New("keyword is required"))
	}
	keywordIsTopic, _ := req.GetArguments()["keywordIsTopic"].(bool)
	keywordInDescription, _ := req.GetArguments()["keywordInDescription"].(bool)
	ownerID := params.GetOptionalInt(req.GetArguments(), "ownerID", 0)
	var pIsPrivate *bool
	isPrivate, ok := req.GetArguments()["isPrivate"].(bool)
	if ok {
		pIsPrivate = new(isPrivate)
	}
	var pIsArchived *bool
	isArchived, ok := req.GetArguments()["isArchived"].(bool)
	if ok {
		pIsArchived = new(isArchived)
	}
	sort, _ := req.GetArguments()["sort"].(string)
	order, _ := req.GetArguments()["order"].(string)
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.SearchRepoOptions{
		Keyword:              keyword,
		KeywordIsTopic:       keywordIsTopic,
		KeywordInDescription: keywordInDescription,
		OwnerID:              ownerID,
		IsPrivate:            pIsPrivate,
		IsArchived:           pIsArchived,
		Sort:                 sort,
		Order:                order,
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	repos, _, err := client.SearchRepos(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("search repos error: %v", err))
	}

	// Convert to compact format
	compactRepos := toCompactRepos(repos)
	return to.TextResult(compactRepos)
}

func ReposFullFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ReposFullFn (full)")
	keyword, ok := req.GetArguments()["keyword"].(string)
	if !ok {
		return to.ErrorResult(errors.New("keyword is required"))
	}
	keywordIsTopic, _ := req.GetArguments()["keywordIsTopic"].(bool)
	keywordInDescription, _ := req.GetArguments()["keywordInDescription"].(bool)
	ownerID := params.GetOptionalInt(req.GetArguments(), "ownerID", 0)
	var pIsPrivate *bool
	isPrivate, ok := req.GetArguments()["isPrivate"].(bool)
	if ok {
		pIsPrivate = new(isPrivate)
	}
	var pIsArchived *bool
	isArchived, ok := req.GetArguments()["isArchived"].(bool)
	if ok {
		pIsArchived = new(isArchived)
	}
	sort, _ := req.GetArguments()["sort"].(string)
	order, _ := req.GetArguments()["order"].(string)
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.SearchRepoOptions{
		Keyword:              keyword,
		KeywordIsTopic:       keywordIsTopic,
		KeywordInDescription: keywordInDescription,
		OwnerID:              ownerID,
		IsPrivate:            pIsPrivate,
		IsArchived:           pIsArchived,
		Sort:                 sort,
		Order:                order,
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	repos, _, err := client.SearchRepos(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("search repos error: %v", err))
	}

	// Return full format
	return to.TextResult(repos)
}

package issue

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
	GetIssueByIndexToolName         = "get_issue_by_index"
	ListRepoIssuesToolName          = "list_repo_issues"
	ListRepoIssuesFullToolName      = "list_repo_issues_full"
	CreateIssueToolName             = "create_issue"
	CreateIssueCommentToolName      = "create_issue_comment"
	EditIssueToolName               = "edit_issue"
	EditIssueCommentToolName        = "edit_issue_comment"
	GetIssueCommentsByIndexToolName = "get_issue_comments_by_index"
)

// CompactIssue represents a minimal issue object for efficient listing
type CompactIssue struct {
	Number     int      `json:"number"`
	Title      string   `json:"title"`
	State      string   `json:"state"`
	Labels     []string `json:"labels,omitempty"`
	Author     string   `json:"author,omitempty"`
	Assignees  []string `json:"assignees,omitempty"`
	Milestone  string   `json:"milestone,omitempty"`
	Comments   int      `json:"comments"`
	Created    string   `json:"created,omitempty"`
	Updated    string   `json:"updated,omitempty"`
	Body       string   `json:"body,omitempty"`
}

// toCompactIssues converts full SDK issues to compact issues
func toCompactIssues(issues []*gitea_sdk.Issue) []CompactIssue {
	result := make([]CompactIssue, len(issues))
	for i, issue := range issues {
		// Extract label names
		var labels []string
		for _, label := range issue.Labels {
			labels = append(labels, label.Name)
		}

		// Get author name
		author := ""
		if issue.Poster != nil {
			author = issue.Poster.UserName
		}

		// Get assignee names
		var assignees []string
		for _, assignee := range issue.Assignees {
			assignees = append(assignees, assignee.UserName)
		}

		// Get milestone title
		milestone := ""
		if issue.Milestone != nil {
			milestone = issue.Milestone.Title
		}

		// Truncate body to preview (first N chars from truncate_compact setting)
		bodyPreview := ""
		if issue.Body != "" {
			if len(issue.Body) > flagPkg.TruncateCompact {
				bodyPreview = issue.Body[:flagPkg.TruncateCompact] + "..."
			} else {
				bodyPreview = issue.Body
			}
		}

		result[i] = CompactIssue{
			Number:    int(issue.Index),
			Title:     issue.Title,
			State:     string(issue.State),
			Labels:    labels,
			Author:    author,
			Assignees: assignees,
			Milestone: milestone,
			Comments:  issue.Comments,
			Created:   issue.Created.Format("2006-01-02 15:04:05"),
			Updated:   issue.Updated.Format("2006-01-02 15:04:05"),
			Body:      bodyPreview,
		}
	}
	return result
}

// truncateIssueBody truncates the body field of issues for full format
func truncateIssueBody(issues []*gitea_sdk.Issue) {
	for _, issue := range issues {
		if issue.Body != "" && len(issue.Body) > flagPkg.TruncateFull {
			issue.Body = issue.Body[:flagPkg.TruncateFull] + "..."
		}
	}
}

var (
	GetIssueByIndexTool = mcp.NewTool(
		GetIssueByIndexToolName,
		mcp.WithDescription("get issue by index"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("repository issue index")),
	)

	ListRepoIssuesTool = mcp.NewTool(
		ListRepoIssuesToolName,
		mcp.WithDescription("List repository issues - compact format, defaults to open state. Use state='all' to include closed issues"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("state", mcp.Description("issue state: open, closed, or all"), mcp.DefaultString("open")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100)),
	)

	ListRepoIssuesFullTool = mcp.NewTool(
		ListRepoIssuesFullToolName,
		mcp.WithDescription("List repository issues - full format, defaults to open state. Use state='all' to include closed issues"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("state", mcp.Description("issue state"), mcp.DefaultString("all")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100)),
	)

	CreateIssueTool = mcp.NewTool(
		CreateIssueToolName,
		mcp.WithDescription("create issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("title", mcp.Required(), mcp.Description("issue title")),
		mcp.WithString("body", mcp.Required(), mcp.Description("issue body")),
	)

	CreateIssueCommentTool = mcp.NewTool(
		CreateIssueCommentToolName,
		mcp.WithDescription("create issue comment"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("repository issue index")),
		mcp.WithString("body", mcp.Required(), mcp.Description("issue comment body")),
	)

	EditIssueTool = mcp.NewTool(
		EditIssueToolName,
		mcp.WithDescription("edit issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("repository issue index")),
		mcp.WithString("title", mcp.Description("issue title"), mcp.DefaultString("")),
		mcp.WithString("body", mcp.Description("issue body content")),
		mcp.WithArray("assignees", mcp.Description("usernames to assign to this issue"), mcp.Items(map[string]any{"type": "string"})),
		mcp.WithNumber("milestone", mcp.Description("milestone number")),
		mcp.WithString("state", mcp.Description("issue state, one of open, closed, all")),
	)

	EditIssueCommentTool = mcp.NewTool(
		EditIssueCommentToolName,
		mcp.WithDescription("edit issue comment"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("commentID", mcp.Required(), mcp.Description("id of issue comment")),
		mcp.WithString("body", mcp.Required(), mcp.Description("issue comment body")),
	)

	GetIssueCommentsByIndexTool = mcp.NewTool(
		GetIssueCommentsByIndexToolName,
		mcp.WithDescription("get issue comment by index"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("repository issue index")),
	)

	// Issue Dependencies
	AddIssueDependencyTool = mcp.NewTool(
		"add_issue_dependency",
		mcp.WithDescription("add a dependency to an issue (marks another issue as blocking this issue)"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue number")),
		mcp.WithNumber("dependency", mcp.Required(), mcp.Description("issue number that blocks this issue")),
	)

	RemoveIssueDependencyTool = mcp.NewTool(
		"remove_issue_dependency",
		mcp.WithDescription("remove a dependency from an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue number")),
		mcp.WithNumber("dependency", mcp.Required(), mcp.Description("issue number dependency to remove")),
	)

	ListIssueDependenciesTool = mcp.NewTool(
		"list_issue_dependencies",
		mcp.WithDescription("list all dependencies (blocking issues) for an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue number")),
		mcp.WithNumber("page", mcp.Description("page number for pagination (optional)")),
		mcp.WithNumber("limit", mcp.Description("page size for pagination (optional)")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetIssueByIndexTool,
		Handler: GetIssueByIndexFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListRepoIssuesTool,
		Handler: ListRepoIssuesFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListRepoIssuesFullTool,
		Handler: ListRepoIssuesFullFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateIssueTool,
		Handler: CreateIssueFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateIssueCommentTool,
		Handler: CreateIssueCommentFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    EditIssueTool,
		Handler: EditIssueFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    EditIssueCommentTool,
		Handler: EditIssueCommentFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetIssueCommentsByIndexTool,
		Handler: GetIssueCommentsByIndexFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    AddIssueDependencyTool,
		Handler: AddIssueDependencyFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    RemoveIssueDependencyTool,
		Handler: RemoveIssueDependencyFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListIssueDependenciesTool,
		Handler: ListIssueDependenciesFn,
	})
}

func GetIssueByIndexFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetIssueByIndexFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issue, _, err := client.GetIssue(owner, repo, index)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get %v/%v/issue/%v err: %v", owner, repo, index, err))
	}

	return to.TextResult(issue)
}

func ListRepoIssuesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoIssuesFn (compact)")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	state, ok := req.GetArguments()["state"].(string)
	if !ok {
		state = "open"
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.ListIssueOption{
		State: gitea_sdk.StateType(state),
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issues, _, err := client.ListRepoIssues(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get %v/%v/issues err: %v", owner, repo, err))
	}
	// Convert to compact format
	compactIssues := toCompactIssues(issues)
	return to.TextResult(compactIssues)
}

func ListRepoIssuesFullFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoIssuesFullFn (full)")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	state, ok := req.GetArguments()["state"].(string)
	if !ok {
		state = "all"
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	opt := gitea_sdk.ListIssueOption{
		State: gitea_sdk.StateType(state),
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issues, _, err := client.ListRepoIssues(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get %v/%v/issues err: %v", owner, repo, err))
	}
	// Truncate body for full format
	truncateIssueBody(issues)
	// Return full format
	return to.TextResult(issues)
}

func CreateIssueFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateIssueFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	title, ok := req.GetArguments()["title"].(string)
	if !ok {
		return to.ErrorResult(errors.New("title is required"))
	}
	body, ok := req.GetArguments()["body"].(string)
	if !ok {
		return to.ErrorResult(errors.New("body is required"))
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issue, _, err := client.CreateIssue(owner, repo, gitea_sdk.CreateIssueOption{
		Title: title,
		Body:  body,
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create %v/%v/issue err: %v", owner, repo, err))
	}

	return to.TextResult(issue)
}

func CreateIssueCommentFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateIssueCommentFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}
	body, ok := req.GetArguments()["body"].(string)
	if !ok {
		return to.ErrorResult(errors.New("body is required"))
	}
	opt := gitea_sdk.CreateIssueCommentOption{
		Body: body,
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issueComment, _, err := client.CreateIssueComment(owner, repo, index, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create %v/%v/issue/%v/comment err: %v", owner, repo, index, err))
	}

	return to.TextResult(issueComment)
}

func EditIssueFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called EditIssueFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}

	opt := gitea_sdk.EditIssueOption{}

	title, ok := req.GetArguments()["title"].(string)
	if ok {
		opt.Title = title
	}
	body, ok := req.GetArguments()["body"].(string)
	if ok {
		opt.Body = &body
	}
	var assignees []string
	if assigneesArg, exists := req.GetArguments()["assignees"]; exists {
		if assigneesSlice, ok := assigneesArg.([]any); ok {
			for _, assignee := range assigneesSlice {
				if assigneeStr, ok := assignee.(string); ok {
					assignees = append(assignees, assigneeStr)
				}
			}
		}
	}
	opt.Assignees = assignees
	if val, exists := req.GetArguments()["milestone"]; exists {
		if milestone, ok := params.ToInt64(val); ok {
			opt.Milestone = &milestone
		}
	}
	state, ok := req.GetArguments()["state"].(string)
	if ok {
		opt.State = new(gitea_sdk.StateType(state))
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issue, _, err := client.EditIssue(owner, repo, index, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit %v/%v/issue/%v err: %v", owner, repo, index, err))
	}

	return to.TextResult(issue)
}

func EditIssueCommentFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called EditIssueCommentFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	commentID, err := params.GetIndex(req.GetArguments(), "commentID")
	if err != nil {
		return to.ErrorResult(err)
	}
	body, ok := req.GetArguments()["body"].(string)
	if !ok {
		return to.ErrorResult(errors.New("body is required"))
	}
	opt := gitea_sdk.EditIssueCommentOption{
		Body: body,
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issueComment, _, err := client.EditIssueComment(owner, repo, commentID, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit %v/%v/issues/comments/%v err: %v", owner, repo, commentID, err))
	}

	return to.TextResult(issueComment)
}

func GetIssueCommentsByIndexFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetIssueCommentsByIndexFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}
	opt := gitea_sdk.ListIssueCommentOptions{}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issue, _, err := client.ListIssueComments(owner, repo, index, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get %v/%v/issues/%v/comments err: %v", owner, repo, index, err))
	}

	return to.TextResult(issue)
}

func AddIssueDependencyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called AddIssueDependencyFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}
	dependency, err := params.GetIndex(req.GetArguments(), "dependency")
	if err != nil {
		return to.ErrorResult(err)
	}

		// Use regular token
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	// Get token from context
	token := flagPkg.Token
	if token == "" {
		return to.ErrorResult(errors.New("token is required for dependency operations"))
	}
	err = createIssueDependencyDirect(client, token, owner, repo, index, dependency)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("add dependency %v to issue %v err: %v", dependency, index, err))
	}

	return to.TextResult("Dependency added successfully")
}

func RemoveIssueDependencyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called RemoveIssueDependencyFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}
	dependency, err := params.GetIndex(req.GetArguments(), "dependency")
	if err != nil {
		return to.ErrorResult(err)
	}

		// Use regular token
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	// Get token from context
	token := flagPkg.Token
	if token == "" {
		return to.ErrorResult(errors.New("token is required for dependency operations"))
	}
	err = removeIssueDependencyDirect(client, token, owner, repo, index, dependency)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("remove dependency %v from issue %v err: %v", dependency, index, err))
	}

	return to.TextResult("Dependency removed successfully")
}

func ListIssueDependenciesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListIssueDependenciesFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	index, err := params.GetIndex(req.GetArguments(), "index")
	if err != nil {
		return to.ErrorResult(err)
	}

	// Optional pagination parameters
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	limit := params.GetOptionalInt(req.GetArguments(), "limit", 10)

	opt := gitea_sdk.ListIssueDependenciesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(limit),
		},
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	dependencies, _, err := client.ListIssueDependencies(owner, repo, index, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list dependencies for issue %v err: %v", index, err))
	}

	return to.TextResult(dependencies)
}

package actions

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ListRepoActionVariablesToolName  = "list_repo_action_variables"
	GetRepoActionVariableToolName    = "get_repo_action_variable"
	CreateRepoActionVariableToolName = "create_repo_action_variable"
	UpdateRepoActionVariableToolName = "update_repo_action_variable"
	DeleteRepoActionVariableToolName = "delete_repo_action_variable"

	ListOrgActionVariablesToolName  = "list_org_action_variables"
	GetOrgActionVariableToolName    = "get_org_action_variable"
	CreateOrgActionVariableToolName = "create_org_action_variable"
	UpdateOrgActionVariableToolName = "update_org_action_variable"
	DeleteOrgActionVariableToolName = "delete_org_action_variable"
)

var (
	ListRepoActionVariablesTool = mcp.NewTool(
		ListRepoActionVariablesToolName,
		mcp.WithDescription("List repository Actions variables"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100), mcp.Min(1)),
	)

	GetRepoActionVariableTool = mcp.NewTool(
		GetRepoActionVariableToolName,
		mcp.WithDescription("Get a repository Actions variable by name"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
	)

	CreateRepoActionVariableTool = mcp.NewTool(
		CreateRepoActionVariableToolName,
		mcp.WithDescription("Create a repository Actions variable"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
		mcp.WithString("value", mcp.Required(), mcp.Description("variable value")),
	)

	UpdateRepoActionVariableTool = mcp.NewTool(
		UpdateRepoActionVariableToolName,
		mcp.WithDescription("Update a repository Actions variable"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
		mcp.WithString("value", mcp.Required(), mcp.Description("new variable value")),
	)

	DeleteRepoActionVariableTool = mcp.NewTool(
		DeleteRepoActionVariableToolName,
		mcp.WithDescription("Delete a repository Actions variable"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
	)

	ListOrgActionVariablesTool = mcp.NewTool(
		ListOrgActionVariablesToolName,
		mcp.WithDescription("List organization Actions variables"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100), mcp.Min(1)),
	)

	GetOrgActionVariableTool = mcp.NewTool(
		GetOrgActionVariableToolName,
		mcp.WithDescription("Get an organization Actions variable by name"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
	)

	CreateOrgActionVariableTool = mcp.NewTool(
		CreateOrgActionVariableToolName,
		mcp.WithDescription("Create an organization Actions variable"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
		mcp.WithString("value", mcp.Required(), mcp.Description("variable value")),
		mcp.WithString("description", mcp.Description("variable description")),
	)

	UpdateOrgActionVariableTool = mcp.NewTool(
		UpdateOrgActionVariableToolName,
		mcp.WithDescription("Update an organization Actions variable"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
		mcp.WithString("value", mcp.Required(), mcp.Description("new variable value")),
		mcp.WithString("description", mcp.Description("new variable description")),
	)

	DeleteOrgActionVariableTool = mcp.NewTool(
		DeleteOrgActionVariableToolName,
		mcp.WithDescription("Delete an organization Actions variable"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("variable name")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{Tool: ListRepoActionVariablesTool, Handler: ListRepoActionVariablesFn})
	Tool.RegisterRead(server.ServerTool{Tool: GetRepoActionVariableTool, Handler: GetRepoActionVariableFn})
	Tool.RegisterWrite(server.ServerTool{Tool: CreateRepoActionVariableTool, Handler: CreateRepoActionVariableFn})
	Tool.RegisterWrite(server.ServerTool{Tool: UpdateRepoActionVariableTool, Handler: UpdateRepoActionVariableFn})
	Tool.RegisterWrite(server.ServerTool{Tool: DeleteRepoActionVariableTool, Handler: DeleteRepoActionVariableFn})

	Tool.RegisterRead(server.ServerTool{Tool: ListOrgActionVariablesTool, Handler: ListOrgActionVariablesFn})
	Tool.RegisterRead(server.ServerTool{Tool: GetOrgActionVariableTool, Handler: GetOrgActionVariableFn})
	Tool.RegisterWrite(server.ServerTool{Tool: CreateOrgActionVariableTool, Handler: CreateOrgActionVariableFn})
	Tool.RegisterWrite(server.ServerTool{Tool: UpdateOrgActionVariableTool, Handler: UpdateOrgActionVariableFn})
	Tool.RegisterWrite(server.ServerTool{Tool: DeleteOrgActionVariableTool, Handler: DeleteOrgActionVariableFn})
}

func ListRepoActionVariablesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoActionVariablesFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)

	query := url.Values{}
	query.Set("page", strconv.Itoa(int(page)))
	query.Set("limit", strconv.Itoa(int(pageSize)))

	var result any
	_, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/actions/variables", url.PathEscape(owner), url.PathEscape(repo)), query, nil, &result)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repo action variables err: %v", err))
	}
	return to.TextResult(result)
}

func GetRepoActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetRepoActionVariableFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	variable, _, err := client.GetRepoActionVariable(owner, repo, name)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get repo action variable err: %v", err))
	}
	return to.TextResult(variable)
}

func CreateRepoActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateRepoActionVariableFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}
	value, ok := req.GetArguments()["value"].(string)
	if !ok || value == "" {
		return to.ErrorResult(errors.New("value is required"))
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	resp, err := client.CreateRepoActionVariable(owner, repo, name, value)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create repo action variable err: %v", err))
	}
	return to.TextResult(map[string]any{"message": "variable created", "status": resp.StatusCode})
}

func UpdateRepoActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called UpdateRepoActionVariableFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}
	value, ok := req.GetArguments()["value"].(string)
	if !ok || value == "" {
		return to.ErrorResult(errors.New("value is required"))
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	resp, err := client.UpdateRepoActionVariable(owner, repo, name, value)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("update repo action variable err: %v", err))
	}
	return to.TextResult(map[string]any{"message": "variable updated", "status": resp.StatusCode})
}

func DeleteRepoActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteRepoActionVariableFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	resp, err := client.DeleteRepoActionVariable(owner, repo, name)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete repo action variable err: %v", err))
	}
	return to.TextResult(map[string]any{"message": "variable deleted", "status": resp.StatusCode})
}

func ListOrgActionVariablesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListOrgActionVariablesFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok || org == "" {
		return to.ErrorResult(errors.New("org is required"))
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	variables, _, err := client.ListOrgActionVariable(org, gitea_sdk.ListOrgActionVariableOption{
		ListOptions: gitea_sdk.ListOptions{Page: int(page), PageSize: int(pageSize)},
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list org action variables err: %v", err))
	}
	return to.TextResult(variables)
}

func GetOrgActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetOrgActionVariableFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok || org == "" {
		return to.ErrorResult(errors.New("org is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	variable, _, err := client.GetOrgActionVariable(org, name)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get org action variable err: %v", err))
	}
	return to.TextResult(variable)
}

func CreateOrgActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateOrgActionVariableFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok || org == "" {
		return to.ErrorResult(errors.New("org is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}
	value, ok := req.GetArguments()["value"].(string)
	if !ok || value == "" {
		return to.ErrorResult(errors.New("value is required"))
	}
	description, _ := req.GetArguments()["description"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	resp, err := client.CreateOrgActionVariable(org, gitea_sdk.CreateOrgActionVariableOption{
		Name:        name,
		Value:       value,
		Description: description,
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create org action variable err: %v", err))
	}
	return to.TextResult(map[string]any{"message": "variable created", "status": resp.StatusCode})
}

func UpdateOrgActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called UpdateOrgActionVariableFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok || org == "" {
		return to.ErrorResult(errors.New("org is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}
	value, ok := req.GetArguments()["value"].(string)
	if !ok || value == "" {
		return to.ErrorResult(errors.New("value is required"))
	}
	description, _ := req.GetArguments()["description"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	resp, err := client.UpdateOrgActionVariable(org, name, gitea_sdk.UpdateOrgActionVariableOption{
		Value:       value,
		Description: description,
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("update org action variable err: %v", err))
	}
	return to.TextResult(map[string]any{"message": "variable updated", "status": resp.StatusCode})
}

func DeleteOrgActionVariableFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteOrgActionVariableFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok || org == "" {
		return to.ErrorResult(errors.New("org is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok || name == "" {
		return to.ErrorResult(errors.New("name is required"))
	}

	_, err := gitea.DoJSON(ctx, "DELETE", fmt.Sprintf("orgs/%s/actions/variables/%s", url.PathEscape(org), url.PathEscape(name)), nil, nil, nil)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete org action variable err: %v", err))
	}
	return to.TextResult(map[string]any{"message": "variable deleted"})
}

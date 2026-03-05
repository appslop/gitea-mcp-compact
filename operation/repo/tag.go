package repo

import (
	"context"
	"errors"
	"fmt"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	CreateTagToolName = "create_tag"
	DeleteTagToolName = "delete_tag"
	GetTagToolName    = "get_tag"
	ListTagsToolName  = "list_tags"
)

var (
	CreateTagTool = mcp.NewTool(
		CreateTagToolName,
		mcp.WithDescription("Create tag"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("tag_name", mcp.Required(), mcp.Description("tag name")),
		mcp.WithString("target", mcp.Description("target commitish"), mcp.DefaultString("")),
		mcp.WithString("message", mcp.Description("tag message"), mcp.DefaultString("")),
	)

	DeleteTagTool = mcp.NewTool(
		DeleteTagToolName,
		mcp.WithDescription("Delete tag"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("tag_name", mcp.Required(), mcp.Description("tag name")),
	)

	GetTagTool = mcp.NewTool(
		GetTagToolName,
		mcp.WithDescription("Get tag"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("tag_name", mcp.Required(), mcp.Description("tag name")),
	)

	ListTagsTool = mcp.NewTool(
		ListTagsToolName,
		mcp.WithDescription("List tags"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(20), mcp.Min(1)),
	)
)

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateTagTool,
		Handler: CreateTagFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteTagTool,
		Handler: DeleteTagFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetTagTool,
		Handler: GetTagFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListTagsTool,
		Handler: ListTagsFn,
	})
}

// To avoid return too many tokens, we need to provide at least information as possible
// llm can call get tag to get more information
type ListTagResult struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	Commit *gitea_sdk.CommitMeta `json:"commit"`
	// message may be a long text, so we should not provide it here
}

func CreateTagFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateTagFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return nil, errors.New("owner is required")
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return nil, errors.New("repo is required")
	}
	tagName, ok := req.GetArguments()["tag_name"].(string)
	if !ok {
		return nil, errors.New("tag_name is required")
	}
	target, _ := req.GetArguments()["target"].(string)
	message, _ := req.GetArguments()["message"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, _, err = client.CreateTag(owner, repo, gitea_sdk.CreateTagOption{
		TagName: tagName,
		Target:  target,
		Message: message,
	})
	if err != nil {
		return nil, fmt.Errorf("create tag error: %v", err)
	}

	return mcp.NewToolResultText("Tag Created"), nil
}

func DeleteTagFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteTagFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return nil, errors.New("owner is required")
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return nil, errors.New("repo is required")
	}
	tagName, ok := req.GetArguments()["tag_name"].(string)
	if !ok {
		return nil, errors.New("tag_name is required")
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteTag(owner, repo, tagName)
	if err != nil {
		return nil, fmt.Errorf("delete tag error: %v", err)
	}

	return to.TextResult("Tag deleted")
}

func GetTagFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetTagFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return nil, errors.New("owner is required")
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return nil, errors.New("repo is required")
	}
	tagName, ok := req.GetArguments()["tag_name"].(string)
	if !ok {
		return nil, errors.New("tag_name is required")
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	tag, _, err := client.GetTag(owner, repo, tagName)
	if err != nil {
		return nil, fmt.Errorf("get tag error: %v", err)
	}

	return to.TextResult(tag)
}

func ListTagsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListTagsFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return nil, errors.New("owner is required")
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return nil, errors.New("repo is required")
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 20)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	tags, _, err := client.ListRepoTags(owner, repo, gitea_sdk.ListRepoTagsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list tags error: %v", err)
	}

	results := make([]ListTagResult, 0, len(tags))
	for _, tag := range tags {
		results = append(results, ListTagResult{
			ID:     tag.ID,
			Name:   tag.Name,
			Commit: tag.Commit,
		})
	}
	return to.TextResult(results)
}

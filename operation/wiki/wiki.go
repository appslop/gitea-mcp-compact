package wiki

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/to"
	"gitea.com/gitea/gitea-mcp/pkg/tool"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var Tool = tool.New()

const (
	ListWikiPagesToolName    = "list_wiki_pages"
	GetWikiPageToolName      = "get_wiki_page"
	GetWikiRevisionsToolName = "get_wiki_revisions"
	CreateWikiPageToolName   = "create_wiki_page"
	UpdateWikiPageToolName   = "update_wiki_page"
	DeleteWikiPageToolName   = "delete_wiki_page"
)

var (
	ListWikiPagesTool = mcp.NewTool(
		ListWikiPagesToolName,
		mcp.WithDescription("List all wiki pages in a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)

	GetWikiPageTool = mcp.NewTool(
		GetWikiPageToolName,
		mcp.WithDescription("Get a wiki page content and metadata"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("pageName", mcp.Required(), mcp.Description("wiki page name")),
	)

	GetWikiRevisionsTool = mcp.NewTool(
		GetWikiRevisionsToolName,
		mcp.WithDescription("Get revisions history of a wiki page"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("pageName", mcp.Required(), mcp.Description("wiki page name")),
	)

	CreateWikiPageTool = mcp.NewTool(
		CreateWikiPageToolName,
		mcp.WithDescription("Create a new wiki page"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("title", mcp.Required(), mcp.Description("wiki page title")),
		mcp.WithString("content_base64", mcp.Required(), mcp.Description("page content, base64 encoded")),
		mcp.WithString("message", mcp.Description("commit message (optional)")),
	)

	UpdateWikiPageTool = mcp.NewTool(
		UpdateWikiPageToolName,
		mcp.WithDescription("Update an existing wiki page"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("pageName", mcp.Required(), mcp.Description("current wiki page name")),
		mcp.WithString("title", mcp.Description("new page title (optional)")),
		mcp.WithString("content_base64", mcp.Required(), mcp.Description("page content, base64 encoded")),
		mcp.WithString("message", mcp.Description("commit message (optional)")),
	)

	DeleteWikiPageTool = mcp.NewTool(
		DeleteWikiPageToolName,
		mcp.WithDescription("Delete a wiki page"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("pageName", mcp.Required(), mcp.Description("wiki page name to delete")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListWikiPagesTool,
		Handler: ListWikiPagesFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetWikiPageTool,
		Handler: GetWikiPageFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetWikiRevisionsTool,
		Handler: GetWikiRevisionsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateWikiPageTool,
		Handler: CreateWikiPageFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    UpdateWikiPageTool,
		Handler: UpdateWikiPageFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteWikiPageTool,
		Handler: DeleteWikiPageFn,
	})
}

func ListWikiPagesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListWikiPagesFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}

	// Use direct HTTP request because SDK does not support yet wikis
	var result any
	_, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/wiki/pages", url.QueryEscape(owner), url.QueryEscape(repo)), nil, nil, &result)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list wiki pages err: %v", err))
	}

	return to.TextResult(result)
}

func GetWikiPageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetWikiPageFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	pageName, ok := req.GetArguments()["pageName"].(string)
	if !ok {
		return to.ErrorResult(errors.New("pageName is required"))
	}

	var result any
	_, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/wiki/page/%s", url.QueryEscape(owner), url.QueryEscape(repo), url.QueryEscape(pageName)), nil, nil, &result)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get wiki page err: %v", err))
	}

	return to.TextResult(result)
}

func GetWikiRevisionsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetWikiRevisionsFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	pageName, ok := req.GetArguments()["pageName"].(string)
	if !ok {
		return to.ErrorResult(errors.New("pageName is required"))
	}

	var result any
	_, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/wiki/revisions/%s", url.QueryEscape(owner), url.QueryEscape(repo), url.QueryEscape(pageName)), nil, nil, &result)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get wiki revisions err: %v", err))
	}

	return to.TextResult(result)
}

func CreateWikiPageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateWikiPageFn")
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
	contentBase64, ok := req.GetArguments()["content_base64"].(string)
	if !ok {
		return to.ErrorResult(errors.New("content_base64 is required"))
	}

	message, _ := req.GetArguments()["message"].(string)
	if message == "" {
		message = fmt.Sprintf("Create wiki page '%s'", title)
	}

	requestBody := map[string]string{
		"title":          title,
		"content_base64": contentBase64,
		"message":        message,
	}

	var result any
	_, err := gitea.DoJSON(ctx, "POST", fmt.Sprintf("repos/%s/%s/wiki/new", url.QueryEscape(owner), url.QueryEscape(repo)), nil, requestBody, &result)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create wiki page err: %v", err))
	}

	return to.TextResult(result)
}

func UpdateWikiPageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called UpdateWikiPageFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	pageName, ok := req.GetArguments()["pageName"].(string)
	if !ok {
		return to.ErrorResult(errors.New("pageName is required"))
	}
	contentBase64, ok := req.GetArguments()["content_base64"].(string)
	if !ok {
		return to.ErrorResult(errors.New("content_base64 is required"))
	}

	requestBody := map[string]string{
		"content_base64": contentBase64,
	}

	// If title is given, use it. Otherwise, keep current page name
	if title, ok := req.GetArguments()["title"].(string); ok && title != "" {
		requestBody["title"] = title
	} else {
		// Utiliser pageName comme fallback pour éviter "unnamed"
		requestBody["title"] = pageName
	}

	if message, ok := req.GetArguments()["message"].(string); ok && message != "" {
		requestBody["message"] = message
	} else {
		requestBody["message"] = fmt.Sprintf("Update wiki page '%s'", pageName)
	}

	var result any
	_, err := gitea.DoJSON(ctx, "PATCH", fmt.Sprintf("repos/%s/%s/wiki/page/%s", url.QueryEscape(owner), url.QueryEscape(repo), url.QueryEscape(pageName)), nil, requestBody, &result)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("update wiki page err: %v", err))
	}

	return to.TextResult(result)
}

func DeleteWikiPageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteWikiPageFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	pageName, ok := req.GetArguments()["pageName"].(string)
	if !ok {
		return to.ErrorResult(errors.New("pageName is required"))
	}

	_, err := gitea.DoJSON(ctx, "DELETE", fmt.Sprintf("repos/%s/%s/wiki/page/%s", url.QueryEscape(owner), url.QueryEscape(repo), url.QueryEscape(pageName)), nil, nil, nil)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete wiki page err: %v", err))
	}

	return to.TextResult(map[string]string{"message": "Wiki page deleted successfully"})
}

package label

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
	ListRepoLabelsToolName     = "list_repo_labels"
	GetRepoLabelToolName       = "get_repo_label"
	CreateRepoLabelToolName    = "create_repo_label"
	EditRepoLabelToolName      = "edit_repo_label"
	DeleteRepoLabelToolName    = "delete_repo_label"
	AddIssueLabelsToolName     = "add_issue_labels"
	ReplaceIssueLabelsToolName = "replace_issue_labels"
	ClearIssueLabelsToolName   = "clear_issue_labels"
	RemoveIssueLabelToolName   = "remove_issue_label"
	ListOrgLabelsToolName      = "list_org_labels"
	CreateOrgLabelToolName     = "create_org_label"
	EditOrgLabelToolName       = "edit_org_label"
	DeleteOrgLabelToolName     = "delete_org_label"
)

var (
	ListRepoLabelsTool = mcp.NewTool(
		ListRepoLabelsToolName,
		mcp.WithDescription("Lists all labels for a given repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100)),
	)

	GetRepoLabelTool = mcp.NewTool(
		GetRepoLabelToolName,
		mcp.WithDescription("Gets a single label by its ID for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("label ID")),
	)

	CreateRepoLabelTool = mcp.NewTool(
		CreateRepoLabelToolName,
		mcp.WithDescription("Creates a new label for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("label name")),
		mcp.WithString("color", mcp.Required(), mcp.Description("label color (hex code, e.g., #RRGGBB)")),
		mcp.WithString("description", mcp.Description("label description")),
	)

	EditRepoLabelTool = mcp.NewTool(
		EditRepoLabelToolName,
		mcp.WithDescription("Edits an existing label in a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("label ID")),
		mcp.WithString("name", mcp.Description("new label name")),
		mcp.WithString("color", mcp.Description("new label color (hex code, e.g., #RRGGBB)")),
		mcp.WithString("description", mcp.Description("new label description")),
	)

	DeleteRepoLabelTool = mcp.NewTool(
		DeleteRepoLabelToolName,
		mcp.WithDescription("Deletes a label from a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("label ID")),
	)

	AddIssueLabelsTool = mcp.NewTool(
		AddIssueLabelsToolName,
		mcp.WithDescription("Adds one or more labels to an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
		mcp.WithArray("labels", mcp.Required(), mcp.Description("array of label IDs to add"), mcp.Items(map[string]any{"type": "number"})),
	)

	ReplaceIssueLabelsTool = mcp.NewTool(
		ReplaceIssueLabelsToolName,
		mcp.WithDescription("Replaces all labels on an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
		mcp.WithArray("labels", mcp.Required(), mcp.Description("array of label IDs to replace with"), mcp.Items(map[string]any{"type": "number"})),
	)

	ClearIssueLabelsTool = mcp.NewTool(
		ClearIssueLabelsToolName,
		mcp.WithDescription("Removes all labels from an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
	)

	RemoveIssueLabelTool = mcp.NewTool(
		RemoveIssueLabelToolName,
		mcp.WithDescription("Removes a single label from an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
		mcp.WithNumber("label_id", mcp.Required(), mcp.Description("label ID to remove")),
	)

	ListOrgLabelsTool = mcp.NewTool(
		ListOrgLabelsToolName,
		mcp.WithDescription("Lists labels defined at organization level"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100)),
	)

	CreateOrgLabelTool = mcp.NewTool(
		CreateOrgLabelToolName,
		mcp.WithDescription("Creates a new label for an organization"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("label name")),
		mcp.WithString("color", mcp.Required(), mcp.Description("label color (hex code, e.g., #RRGGBB)")),
		mcp.WithString("description", mcp.Description("label description")),
		mcp.WithBoolean("exclusive", mcp.Description("whether the label is exclusive"), mcp.DefaultBool(false)),
	)

	EditOrgLabelTool = mcp.NewTool(
		EditOrgLabelToolName,
		mcp.WithDescription("Edits an existing organization label"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("label ID")),
		mcp.WithString("name", mcp.Description("new label name")),
		mcp.WithString("color", mcp.Description("new label color (hex code, e.g., #RRGGBB)")),
		mcp.WithString("description", mcp.Description("new label description")),
		mcp.WithBoolean("exclusive", mcp.Description("whether the label is exclusive")),
	)

	DeleteOrgLabelTool = mcp.NewTool(
		DeleteOrgLabelToolName,
		mcp.WithDescription("Deletes an organization label by ID"),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("label ID")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListRepoLabelsTool,
		Handler: ListRepoLabelsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetRepoLabelTool,
		Handler: GetRepoLabelFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateRepoLabelTool,
		Handler: CreateRepoLabelFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    EditRepoLabelTool,
		Handler: EditRepoLabelFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteRepoLabelTool,
		Handler: DeleteRepoLabelFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    AddIssueLabelsTool,
		Handler: AddIssueLabelsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    ReplaceIssueLabelsTool,
		Handler: ReplaceIssueLabelsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    ClearIssueLabelsTool,
		Handler: ClearIssueLabelsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    RemoveIssueLabelTool,
		Handler: RemoveIssueLabelFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListOrgLabelsTool,
		Handler: ListOrgLabelsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateOrgLabelTool,
		Handler: CreateOrgLabelFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    EditOrgLabelTool,
		Handler: EditOrgLabelFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteOrgLabelTool,
		Handler: DeleteOrgLabelFn,
	})
}

func ListRepoLabelsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoLabelsFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)

	opt := gitea_sdk.ListLabelsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	labels, _, err := client.ListRepoLabels(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list %v/%v/labels err: %v", owner, repo, err))
	}
	return to.TextResult(labels)
}

func GetRepoLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetRepoLabelFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	label, _, err := client.GetRepoLabel(owner, repo, id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get %v/%v/label/%v err: %v", owner, repo, id, err))
	}
	return to.TextResult(label)
}

func CreateRepoLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateRepoLabelFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok {
		return to.ErrorResult(errors.New("name is required"))
	}
	color, ok := req.GetArguments()["color"].(string)
	if !ok {
		return to.ErrorResult(errors.New("color is required"))
	}
	description, _ := req.GetArguments()["description"].(string) // Optional

	opt := gitea_sdk.CreateLabelOption{
		Name:        name,
		Color:       color,
		Description: description,
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	label, _, err := client.CreateLabel(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create %v/%v/label err: %v", owner, repo, err))
	}
	return to.TextResult(label)
}

func EditRepoLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called EditRepoLabelFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}

	opt := gitea_sdk.EditLabelOption{}
	if name, ok := req.GetArguments()["name"].(string); ok {
		opt.Name = new(name)
	}
	if color, ok := req.GetArguments()["color"].(string); ok {
		opt.Color = new(color)
	}
	if description, ok := req.GetArguments()["description"].(string); ok {
		opt.Description = new(description)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	label, _, err := client.EditLabel(owner, repo, id, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit %v/%v/label/%v err: %v", owner, repo, id, err))
	}
	return to.TextResult(label)
}

func DeleteRepoLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteRepoLabelFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok {
		return to.ErrorResult(errors.New("repo is required"))
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteLabel(owner, repo, id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete %v/%v/label/%v err: %v", owner, repo, id, err))
	}
	return to.TextResult("Label deleted successfully")
}

func AddIssueLabelsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called AddIssueLabelsFn")
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
	labelsRaw, ok := req.GetArguments()["labels"].([]any)
	if !ok {
		return to.ErrorResult(errors.New("labels (array of IDs) is required"))
	}
	var labels []int64
	for _, l := range labelsRaw {
		if labelID, ok := params.ToInt64(l); ok {
			labels = append(labels, labelID)
		} else {
			return to.ErrorResult(errors.New("invalid label ID in labels array"))
		}
	}

	opt := gitea_sdk.IssueLabelsOption{
		Labels: labels,
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issueLabels, _, err := client.AddIssueLabels(owner, repo, index, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("add labels to %v/%v/issue/%v err: %v", owner, repo, index, err))
	}
	return to.TextResult(issueLabels)
}

func ReplaceIssueLabelsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ReplaceIssueLabelsFn")
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
	labelsRaw, ok := req.GetArguments()["labels"].([]any)
	if !ok {
		return to.ErrorResult(errors.New("labels (array of IDs) is required"))
	}
	var labels []int64
	for _, l := range labelsRaw {
		if labelID, ok := params.ToInt64(l); ok {
			labels = append(labels, labelID)
		} else {
			return to.ErrorResult(errors.New("invalid label ID in labels array"))
		}
	}

	opt := gitea_sdk.IssueLabelsOption{
		Labels: labels,
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	issueLabels, _, err := client.ReplaceIssueLabels(owner, repo, index, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("replace labels on %v/%v/issue/%v err: %v", owner, repo, index, err))
	}
	return to.TextResult(issueLabels)
}

func ClearIssueLabelsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ClearIssueLabelsFn")
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
	_, err = client.ClearIssueLabels(owner, repo, index)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("clear labels on %v/%v/issue/%v err: %v", owner, repo, index, err))
	}
	return to.TextResult("Labels cleared successfully")
}

func RemoveIssueLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called RemoveIssueLabelFn")
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
	labelID, err := params.GetIndex(req.GetArguments(), "label_id")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteIssueLabel(owner, repo, index, labelID)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("remove label %v from %v/%v/issue/%v err: %v", labelID, owner, repo, index, err))
	}
	return to.TextResult("Label removed successfully")
}

func ListOrgLabelsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListOrgLabelsFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok {
		return to.ErrorResult(errors.New("org is required"))
	}
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)

	opt := gitea_sdk.ListOrgLabelsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	labels, _, err := client.ListOrgLabels(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list %v/labels err: %v", org, err))
	}
	return to.TextResult(labels)
}

func CreateOrgLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateOrgLabelFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok {
		return to.ErrorResult(errors.New("org is required"))
	}
	name, ok := req.GetArguments()["name"].(string)
	if !ok {
		return to.ErrorResult(errors.New("name is required"))
	}
	color, ok := req.GetArguments()["color"].(string)
	if !ok {
		return to.ErrorResult(errors.New("color is required"))
	}
	description, _ := req.GetArguments()["description"].(string)
	exclusive, _ := req.GetArguments()["exclusive"].(bool)

	opt := gitea_sdk.CreateOrgLabelOption{
		Name:        name,
		Color:       color,
		Description: description,
		Exclusive:   exclusive,
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	label, _, err := client.CreateOrgLabel(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create %v/labels err: %v", org, err))
	}
	return to.TextResult(label)
}

func EditOrgLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called EditOrgLabelFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok {
		return to.ErrorResult(errors.New("org is required"))
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}

	opt := gitea_sdk.EditOrgLabelOption{}
	if name, ok := req.GetArguments()["name"].(string); ok {
		opt.Name = new(name)
	}
	if color, ok := req.GetArguments()["color"].(string); ok {
		opt.Color = new(color)
	}
	if description, ok := req.GetArguments()["description"].(string); ok {
		opt.Description = new(description)
	}
	if exclusive, ok := req.GetArguments()["exclusive"].(bool); ok {
		opt.Exclusive = new(exclusive)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	label, _, err := client.EditOrgLabel(org, id, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit %v/labels/%v err: %v", org, id, err))
	}
	return to.TextResult(label)
}

func DeleteOrgLabelFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteOrgLabelFn")
	org, ok := req.GetArguments()["org"].(string)
	if !ok {
		return to.ErrorResult(errors.New("org is required"))
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteOrgLabel(org, id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete %v/labels/%v err: %v", org, id, err))
	}
	return to.TextResult("Label deleted successfully")
}

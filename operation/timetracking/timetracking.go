// Package timetracking provides MCP tools for Gitea time tracking operations
package timetracking

import (
	"context"
	"errors"
	"fmt"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"
	"gitea.com/gitea/gitea-mcp/pkg/tool"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var Tool = tool.New()

const (
	// Stopwatch tools
	StartStopwatchToolName   = "start_stopwatch"
	StopStopwatchToolName    = "stop_stopwatch"
	DeleteStopwatchToolName  = "delete_stopwatch"
	GetMyStopwatchesToolName = "get_my_stopwatches"

	// Tracked time tools
	ListTrackedTimesToolName  = "list_tracked_times"
	AddTrackedTimeToolName    = "add_tracked_time"
	DeleteTrackedTimeToolName = "delete_tracked_time"
	ListRepoTimesToolName     = "list_repo_times"
	GetMyTimesToolName        = "get_my_times"
)

var (
	// Stopwatch tools
	StartStopwatchTool = mcp.NewTool(
		StartStopwatchToolName,
		mcp.WithDescription("Start a stopwatch on an issue to track time spent"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
	)

	StopStopwatchTool = mcp.NewTool(
		StopStopwatchToolName,
		mcp.WithDescription("Stop a running stopwatch on an issue and record the tracked time"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
	)

	DeleteStopwatchTool = mcp.NewTool(
		DeleteStopwatchToolName,
		mcp.WithDescription("Delete/cancel a running stopwatch without recording time"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
	)

	GetMyStopwatchesTool = mcp.NewTool(
		GetMyStopwatchesToolName,
		mcp.WithDescription("Get all currently running stopwatches for the authenticated user"),
	)

	// Tracked time tools
	ListTrackedTimesTool = mcp.NewTool(
		ListTrackedTimesToolName,
		mcp.WithDescription("List tracked times for a specific issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100)),
	)

	AddTrackedTimeTool = mcp.NewTool(
		AddTrackedTimeToolName,
		mcp.WithDescription("Manually add tracked time to an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
		mcp.WithNumber("time", mcp.Required(), mcp.Description("time to add in seconds")),
	)

	DeleteTrackedTimeTool = mcp.NewTool(
		DeleteTrackedTimeToolName,
		mcp.WithDescription("Delete a tracked time entry from an issue"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("issue index")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("tracked time entry ID")),
	)

	ListRepoTimesTool = mcp.NewTool(
		ListRepoTimesToolName,
		mcp.WithDescription("List all tracked times for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("pageSize", mcp.Description("page size"), mcp.DefaultNumber(100)),
	)

	GetMyTimesTool = mcp.NewTool(
		GetMyTimesToolName,
		mcp.WithDescription("Get all tracked times for the authenticated user"),
	)
)

func init() {
	// Stopwatch tools
	Tool.RegisterWrite(server.ServerTool{Tool: StartStopwatchTool, Handler: StartStopwatchFn})
	Tool.RegisterWrite(server.ServerTool{Tool: StopStopwatchTool, Handler: StopStopwatchFn})
	Tool.RegisterWrite(server.ServerTool{Tool: DeleteStopwatchTool, Handler: DeleteStopwatchFn})
	Tool.RegisterRead(server.ServerTool{Tool: GetMyStopwatchesTool, Handler: GetMyStopwatchesFn})

	// Tracked time tools
	Tool.RegisterRead(server.ServerTool{Tool: ListTrackedTimesTool, Handler: ListTrackedTimesFn})
	Tool.RegisterWrite(server.ServerTool{Tool: AddTrackedTimeTool, Handler: AddTrackedTimeFn})
	Tool.RegisterWrite(server.ServerTool{Tool: DeleteTrackedTimeTool, Handler: DeleteTrackedTimeFn})
	Tool.RegisterRead(server.ServerTool{Tool: ListRepoTimesTool, Handler: ListRepoTimesFn})
	Tool.RegisterRead(server.ServerTool{Tool: GetMyTimesTool, Handler: GetMyTimesFn})
}

// Stopwatch handler functions

func StartStopwatchFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called StartStopwatchFn")
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
	_, err = client.StartIssueStopWatch(owner, repo, index)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("start stopwatch on %s/%s#%d err: %v", owner, repo, index, err))
	}
	return to.TextResult(fmt.Sprintf("Stopwatch started on issue %s/%s#%d", owner, repo, index))
}

func StopStopwatchFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called StopStopwatchFn")
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
	_, err = client.StopIssueStopWatch(owner, repo, index)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("stop stopwatch on %s/%s#%d err: %v", owner, repo, index, err))
	}
	return to.TextResult(fmt.Sprintf("Stopwatch stopped on issue %s/%s#%d - time recorded", owner, repo, index))
}

func DeleteStopwatchFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteStopwatchFn")
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
	_, err = client.DeleteIssueStopwatch(owner, repo, index)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete stopwatch on %s/%s#%d err: %v", owner, repo, index, err))
	}
	return to.TextResult(fmt.Sprintf("Stopwatch deleted/cancelled on issue %s/%s#%d", owner, repo, index))
}

func GetMyStopwatchesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetMyStopwatchesFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	stopwatches, _, err := client.ListMyStopwatches(gitea_sdk.ListStopwatchesOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get stopwatches err: %v", err))
	}
	if len(stopwatches) == 0 {
		return to.TextResult("No active stopwatches")
	}
	return to.TextResult(stopwatches)
}

// Tracked time handler functions

func ListTrackedTimesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListTrackedTimesFn")
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
	page := params.GetOptionalInt(req.GetArguments(), "page", 1)
	pageSize := params.GetOptionalInt(req.GetArguments(), "pageSize", 100)
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	times, _, err := client.ListIssueTrackedTimes(owner, repo, index, gitea_sdk.ListTrackedTimesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list tracked times for %s/%s#%d err: %v", owner, repo, index, err))
	}
	if len(times) == 0 {
		return to.TextResult(fmt.Sprintf("No tracked times for issue %s/%s#%d", owner, repo, index))
	}
	return to.TextResult(times)
}

func AddTrackedTimeFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called AddTrackedTimeFn")
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

	timeSeconds, err := params.GetIndex(req.GetArguments(), "time")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	trackedTime, _, err := client.AddTime(owner, repo, index, gitea_sdk.AddTimeOption{
		Time: timeSeconds,
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("add tracked time to %s/%s#%d err: %v", owner, repo, index, err))
	}
	return to.TextResult(trackedTime)
}

func DeleteTrackedTimeFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DeleteTrackedTimeFn")
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
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteTime(owner, repo, index, id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete tracked time %d from %s/%s#%d err: %v", id, owner, repo, index, err))
	}
	return to.TextResult(fmt.Sprintf("Tracked time entry %d deleted from issue %s/%s#%d", id, owner, repo, index))
}

func ListRepoTimesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoTimesFn")
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
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	times, _, err := client.ListRepoTrackedTimes(owner, repo, gitea_sdk.ListTrackedTimesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     int(page),
			PageSize: int(pageSize),
		},
	})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repo tracked times for %s/%s err: %v", owner, repo, err))
	}
	if len(times) == 0 {
		return to.TextResult(fmt.Sprintf("No tracked times for repository %s/%s", owner, repo))
	}
	return to.TextResult(times)
}

func GetMyTimesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetMyTimesFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	times, _, err := client.ListMyTrackedTimes(gitea_sdk.ListTrackedTimesOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get tracked times err: %v", err))
	}
	if len(times) == 0 {
		return to.TextResult("No tracked times found")
	}
	return to.TextResult(times)
}

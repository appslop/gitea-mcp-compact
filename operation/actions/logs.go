package actions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	GetRepoActionJobLogPreviewToolName = "get_repo_action_job_log_preview"
	DownloadRepoActionJobLogToolName   = "download_repo_action_job_log"
)

var (
	GetRepoActionJobLogPreviewTool = mcp.NewTool(
		GetRepoActionJobLogPreviewToolName,
		mcp.WithDescription("Get a repository Actions job log preview (tail/limited for chat-friendly output)"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("job_id", mcp.Required(), mcp.Description("job ID")),
		mcp.WithNumber("tail_lines", mcp.Description("number of lines from the end of the log"), mcp.DefaultNumber(200), mcp.Min(1)),
		mcp.WithNumber("max_bytes", mcp.Description("max bytes to return"), mcp.DefaultNumber(65536), mcp.Min(1024)),
	)

	DownloadRepoActionJobLogTool = mcp.NewTool(
		DownloadRepoActionJobLogToolName,
		mcp.WithDescription("Download a repository Actions job log to a file on the MCP server filesystem"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("job_id", mcp.Required(), mcp.Description("job ID")),
		mcp.WithString("output_path", mcp.Description("optional output file path; if omitted, uses ~/.gitea-mcp/artifacts/actions-logs/...")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{Tool: GetRepoActionJobLogPreviewTool, Handler: GetRepoActionJobLogPreviewFn})
	Tool.RegisterRead(server.ServerTool{Tool: DownloadRepoActionJobLogTool, Handler: DownloadRepoActionJobLogFn})
}

func logPaths(owner, repo string, jobID int64) []string {
	// Primary candidate endpoints, plus a few commonly-seen variants across versions.
	// We try these in order; 404/405 falls through.
	return []string{
		fmt.Sprintf("repos/%s/%s/actions/jobs/%d/logs", url.PathEscape(owner), url.PathEscape(repo), jobID),
		fmt.Sprintf("repos/%s/%s/actions/jobs/%d/log", url.PathEscape(owner), url.PathEscape(repo), jobID),
		fmt.Sprintf("repos/%s/%s/actions/tasks/%d/log", url.PathEscape(owner), url.PathEscape(repo), jobID),
		fmt.Sprintf("repos/%s/%s/actions/task/%d/log", url.PathEscape(owner), url.PathEscape(repo), jobID),
	}
}

func fetchJobLogBytes(ctx context.Context, owner, repo string, jobID int64) ([]byte, string, error) {
	var lastErr error
	for _, p := range logPaths(owner, repo, jobID) {
		b, _, err := gitea.DoBytes(ctx, "GET", p, nil, nil, "text/plain")
		if err == nil {
			return b, p, nil
		}
		lastErr = err
		var httpErr *gitea.HTTPError
		if errors.As(err, &httpErr) && (httpErr.StatusCode == http.StatusNotFound || httpErr.StatusCode == http.StatusMethodNotAllowed) {
			continue
		}
		return nil, p, err
	}
	return nil, "", lastErr
}

func tailByLines(data []byte, tailLines int) []byte {
	if tailLines <= 0 || len(data) == 0 {
		return data
	}
	// Find the start index of the last N lines by scanning backwards.
	lines := 0
	i := len(data) - 1
	for i >= 0 {
		if data[i] == '\n' {
			lines++
			if lines > tailLines {
				return data[i+1:]
			}
		}
		i--
	}
	return data
}

func limitBytes(data []byte, maxBytes int) ([]byte, bool) {
	if maxBytes <= 0 {
		return data, false
	}
	if len(data) <= maxBytes {
		return data, false
	}
	// Keep the tail so the most recent log content is preserved.
	return data[len(data)-maxBytes:], true
}

func GetRepoActionJobLogPreviewFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetRepoActionJobLogPreviewFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	jobID, err := params.GetIndex(req.GetArguments(), "job_id")
	if err != nil || jobID <= 0 {
		return to.ErrorResult(errors.New("job_id is required"))
	}
	tailLines := int(params.GetOptionalInt(req.GetArguments(), "tail_lines", 200))
	maxBytes := int(params.GetOptionalInt(req.GetArguments(), "max_bytes", 65536))
	raw, usedPath, err := fetchJobLogBytes(ctx, owner, repo, jobID)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get job log err: %v", err))
	}

	tailed := tailByLines(raw, tailLines)
	limited, truncated := limitBytes(tailed, maxBytes)

	return to.TextResult(map[string]any{
		"endpoint":   usedPath,
		"job_id":     jobID,
		"bytes":      len(raw),
		"tail_lines": tailLines,
		"max_bytes":  maxBytes,
		"truncated":  truncated,
		"log":        string(limited),
	})
}

func DownloadRepoActionJobLogFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called DownloadRepoActionJobLogFn")
	owner, ok := req.GetArguments()["owner"].(string)
	if !ok || owner == "" {
		return to.ErrorResult(errors.New("owner is required"))
	}
	repo, ok := req.GetArguments()["repo"].(string)
	if !ok || repo == "" {
		return to.ErrorResult(errors.New("repo is required"))
	}
	jobID, err := params.GetIndex(req.GetArguments(), "job_id")
	if err != nil || jobID <= 0 {
		return to.ErrorResult(errors.New("job_id is required"))
	}
	outputPath, _ := req.GetArguments()["output_path"].(string)

	raw, usedPath, err := fetchJobLogBytes(ctx, owner, repo, jobID)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("download job log err: %v", err))
	}

	if outputPath == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = os.TempDir()
		}
		outputPath = filepath.Join(home, ".gitea-mcp", "artifacts", "actions-logs", owner, repo, fmt.Sprintf("%d.log", jobID))
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return to.ErrorResult(fmt.Errorf("create output dir err: %v", err))
	}
	if err := os.WriteFile(outputPath, raw, 0o600); err != nil {
		return to.ErrorResult(fmt.Errorf("write log file err: %v", err))
	}

	return to.TextResult(map[string]any{
		"endpoint": usedPath,
		"job_id":   jobID,
		"path":     outputPath,
		"bytes":    len(raw),
	})
}

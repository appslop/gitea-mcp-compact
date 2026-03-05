package issue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	gitea_sdk "code.gitea.io/sdk/gitea"
	flagPkg "gitea.com/gitea/gitea-mcp/pkg/flag"
)

// createIssueDependencyDirect makes a direct HTTP call to add a dependency
// This bypasses the gitea-go-sdk limitation where IssueMeta only has Index field
func createIssueDependencyDirect(client *gitea_sdk.Client, token, owner, repo string, index, dependency int64) error {
	body := map[string]interface{}{
		"index": dependency,
		"owner": owner,
		"repo":  repo,
	}

	jsonBody, _ := json.Marshal(body)
	url := flagPkg.Host + fmt.Sprintf("/api/v1/repos/%s/%s/issues/%d/dependencies", owner, repo, index)

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// removeIssueDependencyDirect makes a direct HTTP call to remove a dependency
func removeIssueDependencyDirect(client *gitea_sdk.Client, token, owner, repo string, index, dependency int64) error {
	body := map[string]interface{}{
		"index": dependency,
		"owner": owner,
		"repo":  repo,
	}

	jsonBody, _ := json.Marshal(body)
	url := flagPkg.Host + fmt.Sprintf("/api/v1/repos/%s/%s/issues/%d/dependencies", owner, repo, index)

	req, err := http.NewRequest("DELETE", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

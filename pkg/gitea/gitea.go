package gitea

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"
	mcpContext "gitea.com/gitea/gitea-mcp/pkg/context"
	"gitea.com/gitea/gitea-mcp/pkg/flag"
)

func NewClient(token string) (*gitea.Client, error) {
	var transport http.Transport
	if flag.Insecure {
		transport = http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	} else {
		// Copy default transport values for a fresh transport
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if ok {
			tmp := *defaultTransport // Copy the struct
			transport = tmp
		} else {
			transport = http.Transport{}
		}
	}

	httpClient := &http.Client{
		Transport: &transport,
	}

	opts := []gitea.ClientOption{
		gitea.SetToken(token),
		gitea.SetHTTPClient(httpClient),
	}
	if flag.Debug {
		opts = append(opts, gitea.SetDebugMode())
	}
	client, err := gitea.NewClient(flag.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("create gitea client err: %w", err)
	}

	// Set user agent for the client
	client.SetUserAgent("gitea-mcp-server/" + flag.Version)
	return client, nil
}

func ClientFromContext(ctx context.Context) (*gitea.Client, error) {
	token, ok := ctx.Value(mcpContext.TokenContextKey).(string)
	if !ok {
		token = flag.Token
	}
	return NewClient(token)
}

// UserClientFromContext returns a Gitea client using the user token (elevated permissions).
// This is used for privileged operations that require special access beyond the regular agent token.
// Token precedence: HTTP X-User-Token header > --user-token flag > GITEA_USER_ACCESS_TOKEN env var
func UserClientFromContext(ctx context.Context) (*gitea.Client, error) {
	token, ok := ctx.Value(mcpContext.UserTokenContextKey).(string)
	if !ok {
		token = flag.UserToken
	}
	return NewClient(token)
}

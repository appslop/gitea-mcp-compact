package gitea

import (
	"context"
	"testing"

	mcpContext "gitea.com/gitea/gitea-mcp/pkg/context"
	"gitea.com/gitea/gitea-mcp/pkg/flag"
)

func TestTokenFromContext(t *testing.T) {
	orig := flag.Token
	defer func() { flag.Token = orig }()

	flag.Token = "flag-token"

	t.Run("context token wins", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), mcpContext.TokenContextKey, "ctx-token")
		if got := tokenFromContext(ctx); got != "ctx-token" {
			t.Fatalf("tokenFromContext() = %q, want %q", got, "ctx-token")
		}
	})

	t.Run("fallback to flag token", func(t *testing.T) {
		ctx := context.Background()
		if got := tokenFromContext(ctx); got != "flag-token" {
			t.Fatalf("tokenFromContext() = %q, want %q", got, "flag-token")
		}
	})
}

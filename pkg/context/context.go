package context

type contextKey string

const (
	TokenContextKey     = contextKey("token")
	UserTokenContextKey = contextKey("userToken")
)

package flag

var (
	Host             string
	Port             int
	Token            string
	UserToken        string
	Version          string
	Mode             string
	CreateRepoAsUser bool
	TruncateCompact  int
	TruncateFull     int

	Insecure bool
	ReadOnly bool
	Debug    bool
)

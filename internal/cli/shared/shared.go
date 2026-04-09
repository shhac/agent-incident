package shared

// GlobalFlags holds persistent flags available to all commands.
type GlobalFlags struct {
	Org     string
	APIKey  string
	Format  string
	Timeout int
}

// GlobalsFunc is the signature for the globals accessor passed to domain Register functions.
type GlobalsFunc = func() *GlobalFlags

package sam3

// Options represents a map of configuration options
type Options map[string]string

// AsList returns options as a list of strings
func (opts Options) AsList() (ls []string) {
	for k, v := range opts {
		ls = append(ls, k+"="+v)
	}
	return
}

// Option is a functional option for SAMEmit
type Option func(*SAMEmit) error

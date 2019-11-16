package ctx

import "os"

// Env represents a set of environment variables
type Env map[string]string

// NewEnv returns a new set of environment variables
func NewEnv() Env {
	return map[string]string{}
}

// Set sets an environment variable
func (e Env) Set(key, val string) {
	e[key] = val
}

// Get retrieves an environment variable
func (e Env) Get(key string) (string, bool) {
	val, ok := e[key]
	return val, ok
}

// Has returns whether an environment variable is set
func (e Env) Has(key string) bool {
	_, ok := e[key]
	return ok
}

// Expand does variable expansion for strings provided by replacing `${var}`
// and `$var` with values set in Env. Undefined variables replaced by an
// empty string.
func (e Env) Expand(s string) string {
	return os.Expand(s, func(key string) string {
		if val, ok := e.Get(key); ok {
			return val
		}
		return ""
	})
}

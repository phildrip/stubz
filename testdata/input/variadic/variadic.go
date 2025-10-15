package variadic

import "context"

// VariadicInterface has methods with variadic parameters
type VariadicInterface interface {
	// Single variadic parameter
	Log(level string, args ...interface{}) error

	// Variadic parameter with other parameters before it
	Execute(ctx context.Context, command string, args ...string) (int, error)

	// Non-variadic method for comparison
	Simple(a, b int) int
}

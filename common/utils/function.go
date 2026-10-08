package utils

import (
	"reflect"
	"runtime"
	"strings"
)

/*
**
Why any is Better
  - Identical under the hood: The Go language defines any directly as type any = interface{}. They behave exactly the same way during execution.
  - Better readability: Writing map[string]any is much easier to scan and read than map[string]interface{}.
  - Idiomatic code: The Go team introduced any alongside generics. Popular static analysis tools
    and linter configurations like golangci-lint now explicitly flag interface{} and suggest replacing it with any.

When to use interface{}Legacy codebases:
  - If you are working on an old codebase that must compile with versions older than Go 1.18.
  - Defining actual methods: You still need the interface keyword when defining an interface
    that requires behavior (e.g., type Reader interface { Read() })

**
*/
func GetFunctionName(i any) string {
	packageFnName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	if packageFnName == "" {
		return ""
	}
	packagesFuncs := strings.Split(packageFnName, ".")
	if len(packagesFuncs) < 1 {
		return ""
	}
	return packagesFuncs[len(packagesFuncs)-1]
}

// Extract struct name from a pointer or struct value (may include package name)
func GetStructName(i any) string {
	t := reflect.TypeOf(i)
	// if the argument is a pointer, get the element type
	if t.Kind() == reflect.Pointer {
		return t.Elem().Name()
	}
	return t.Name()
}

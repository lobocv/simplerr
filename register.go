package simplerr

import "fmt"

var (
	registry *Registry
)

// init initializes the default registry with some convenient defaults
func init() {
	registry = NewRegistry()
	for code, description := range defaultErrorCodes {
		registry.codeDescriptions[code] = description
	}
}

// GetRegistry gets the currently set registry
func GetRegistry() *Registry {
	return registry
}

// SetRegistry sets the default error registry
func SetRegistry(r *Registry) {
	registry = r
}

// Registry is a registry of information on how to handle and serve simple errors
type Registry struct {
	codeDescriptions map[Code]string
}

// NewRegistry creates a new registry without any defaults
func NewRegistry() *Registry {
	return &Registry{
		codeDescriptions: map[Code]string{},
	}
}

// RegisterErrorCode registers custom error codes in the registry. This call will panic if the error code is already registered.
// Error codes 0-99 are reserved for simplerr.
// This method should be called early on application startup.
func (r *Registry) RegisterErrorCode(code Code, description string) {
	if _, exists := r.codeDescriptions[code]; exists {
		panic("error code %s:%d already registered")
	}

	if code < NumberOfReservedCodes {
		panic(fmt.Sprintf("SimpleError codes 0 to %d are reserved.", NumberOfReservedCodes-1))
	}
	r.codeDescriptions[code] = description
}

// ErrorCodes returns a copy of the registered error codes and their descriptions
func (r *Registry) ErrorCodes() map[Code]string {
	codes := make(map[Code]string, len(defaultErrorCodes))
	for k, v := range r.codeDescriptions {
		codes[k] = v
	}
	return codes
}

// CodeDescription returns the description of the error code
func (r *Registry) CodeDescription(c Code) string {
	return r.codeDescriptions[c]
}

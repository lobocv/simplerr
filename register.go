package simplerr

import "fmt"

type Code int

var (
	registry *Registry
)

// init initializes the default registry with some convenient defaults
func init() {
	registry = NewRegistry()
	for code, description := range defaultErrorCodes {
		registry.codeDescriptions[code] = description
	}
	registry.conversions = append(registry.conversions, defaultErrorConversions...)
}

// SetRegistry sets the default error registry
func SetRegistry(r *Registry) {
	registry = r
}

// ErrorConversion is a function that converts an `error` into a  simple SimpleError{}
type ErrorConversion func(err error) *SimpleError

// Registry is a registry of information on how to handle and serve simple errors
type Registry struct {
	codeDescriptions map[Code]string
	conversions      []ErrorConversion
}

// NewRegistry creates a new registry without any defaults
func NewRegistry() *Registry {
	return &Registry{
		conversions:      []ErrorConversion{},
		codeDescriptions: map[Code]string{},
	}
}

// RegisterErrorCode registers custom error codes in the registry. This call may panic if a reserved error code is used.
// This method should be called early on application startup.
func (r *Registry) RegisterErrorCode(code Code, description string) {
	if code < NumberOfReservedCodes {
		panic(fmt.Sprintf("SimpleError codes 0 to %d are reserved.", NumberOfReservedCodes-1))
	}

	if _, exists := defaultErrorCodes[code]; exists {
		panic("error code %s:%d already registered")
	}

	r.codeDescriptions[code] = description
}

// RegisterErrorConversions registers an error conversion function
func (r *Registry) RegisterErrorConversions(funcs ...ErrorConversion) {
	r.conversions = append(r.conversions, funcs...)
}

// ErrorCodes returns a copy of the registered error codes and their descriptions
func (r *Registry) ErrorCodes() map[Code]string {
	codes := make(map[Code]string, len(defaultErrorCodes))
	for k, v := range r.codeDescriptions {
		codes[k] = v
	}
	return codes
}

func (r *Registry) CodeDescription(c Code) string {
	return r.codeDescriptions[c]
}

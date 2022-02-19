[![Go Report Card](https://goreportcard.com/badge/github.com/lobocv/simplerr)](https://goreportcard.com/report/github.com/lobocv/simplerr)
[<img src="https://img.shields.io/github/license/lobocv/simplerr">](https://img.shields.io/github/license/lobocv/simplerr)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-100%25-brightgreen.svg?longCache=true&style=flat)</a>
![Build Status](https://github.com/lobocv/simplerr/actions/workflows/build.yaml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/lobocv/simplerr.svg)](https://pkg.go.dev/github.com/lobocv/simplerr)

# Simplerr
<p align="center"><img src="gopher.png" width="250"></p>

Simplerr provides a simple and more powerful Go error handling experience by providing an alternative `error` 
implementation, the [SimpleError](https://pkg.go.dev/github.com/lobocv/simplerr#SimpleError). Simplerr was designed to
be convenient and highly configurable. The main goals of Simplerr is to reduce boilerplate and make error handling
and debugging easier.


# Features

The `SimpleError` allows you to easily:

- Apply an error code to any error. Choose from a list of standard codes or [register](https://pkg.go.dev/github.com/lobocv/simplerr#Registry) your own.
- Register `func(err) *SimpleError` conversion functions to easily convert to `SimpleErrors` using `Convert()`.
- Automatically translate `simplerr` (including custom codes) error codes to other standardized codes such as `HTTP/gRPC`.
- Attach key-value pairs to errors to be used with structured loggers.
- Attach custom attributes similar to the `context` package.
- Capture stack traces at the point the error is raised.
- Mark errors as `silent` so they can be skipped by logging middleware.
- Mark errors as `benign` so they can be logged less severely by logging middleware.

A complete list of standard error codes can be found [here](https://github.com/lobocv/simplerr/blob/master/codes.go).

# Basic usage

## Creating errors
Errors can be created with `New(format string, args... interface{})`, which works similar to `fmt.Errorf` but instead
returns a `*SimplerError`. You can then chain mutations onto the error to add additional information.

```go
userID := 123
companyID := 456
err := simplerr.New("user %d does not exist in company %d", userID, companyID).
	Code(CodeNotFound).
	Aux("user_id", userID, "company_id", companyID)
```

In the above example, a new error is created and set to error code `CodeNotFound`. We have also attached auxiliary
key-value pair information to the error that we can extract later on when we decide to handle or log the error.

Errors can also be wrapped with the `Wrap(err error)` and `Wrapf(err error, format string, args... []interface{})` functions:

```go
func GetUser(userID int) (*User, error) {
    user, err := db.GetUser(userID)
    if err != nil {
        serr = simplerr.Wrapf(err, "failed to get user with id = %d", userID).Aux("user_id", userID)
        if errors.Is(err, sql.ErrNoRows) {
            serr.Code(CodeNotFound)   
        }
        return serr
    }
}
```

## Automatic error conversion:

The above example where we manually check for `sql.ErrNoRows` can be cleaned up further by globally registering an error 
conversion function:

```go

func main() {
    // Register a conversion function for sql.ErrNoRows to be converted to CodeNotFound
    r := simplerr.GetRegistry()
    r.RegisterErrorConversions(func(err error) simplerr.*SimpleError {
        if errors.Is(err, sql.ErrNoRows) {
            return simplerr.Wrap(err).Code(CodeNotFound)
        }
        return nil
    })
    /// ...
}
```
and using `Convert()`:
```go
func GetUser(userID int) (*User, error) {
    user, err := db.GetUser(userID)
    if err != nil {
        return nil, simplerr.Convert(err).
    		       Message("failed to get user with id = %d", userID).
    		       Aux("user_id", userID)
    }
    return user, nil
```

Calling `Convert()` will run the error through all registered conversion functions and 
use the first result that returns a non-nil value. In the above example, the error code
will be set to `CodeNotFound`.

## Attaching Custom Attributes to Errors

Simplerr lets you define and detect your own custom attributes on errors. This works similarly to the `context` package.
An attribute is attached to an error using the `Attr()` mutator and can be retrieved using the `GetAttribute()` function,
which finds the first match of the attribute key in the error chain.

It is highly recommended that a custom type be used as the key in order to prevent naming collisions of attributes. 
The following example defines a `NotRetryable` attribute and attaches it on an error where a unique constraint is violated,
this indicates that the error should be exempt by any retry mechanism. 


```go

// Define a custom type so we don't get naming collisions for value == 1
type ErrorAttribute int

// Define a specific key for the attribute
const NotRetryable = ErrorAttribute(1)

// Attach the `NotRetryable` attribute on the error
serr := simplerr.New("user with that email already exists").
	Code(CodeConstraintViolated).
	Attr(NotRetryable, true)

// Get the value of the NotRetryable attribute
isRetryable := simplerr.GetAttribute(err, NotRetryable).(bool)
// isRetryable == true
```

## Detecting errors

`SimpleError` implements the `Unwrap` method so it can be used with the standard library
`errors.Is()` and `errors.As()` functions. However, the ability to use error codes makes
abstracting and detecting errors much simpler. Instead of looking for a specific error, `simplerr`
allows you to search for the **kind** of error by looking for an error code:

```go
func GetUserSettings(userID int) (*Settings, error) {
    settings, err := db.GetSettings(userID)
    if err != nil {
        // If the settings do not exist, return defaults
        if simplerr.HasErrorCode(CodeNotFound) {
            return defaultSettings(), nil
        }
		
        serr := simplerr.Wrapf(err, "failed to get settings for user with id = %d", userID).
                         Aux("user_id", userID)
        return nil, serr
    }
	
    return settings, nil
}
```

The alternatives would be to use `errors.Is(err, sql.ErrNoRows)` directly and leak an implementation
detail of the persistence layer or to define a custom error that the persistence layer would need
to return in place of `sql.ErrNoRows`. 

# Error Handling 

`SimpleErrors` were designed to be handled. The [ecosystem](https://github.com/lobocv/simplerr/tree/master/ecosystem)
package provides packages to assist with error handling for different applications. Designing your own handlers is as
simple as detecting the `SimpleError` and reacting to it's attributes.

## Detecting Errors

To detect a specific error code, you can use `HasErrorCode(err error, c Code)`. If you want to look for several different
error codes, use `HasErrorCodes(err error, codes... Code)`, which returns the first of the provided error codes that is 
detected, and a boolean for whether anything was detected.


## Logging SimpleErrors

One of the objective to `simplerr` is to reduce the need to log the errors manually at the sight in which they are raised,
and instead, log errors in a procedural way in a middleware layer. While this is possible with standard library errors,
there is a lack of control when dealing only with the simple `string`-backed error implementation.

### Logging with Structured Loggers

It is good practice to use structured logging to improve observability. However, the standard library error does
not allow for attaching and retrieving key-value pairs on errors. With `simplerr` you can retrieve a superset of all attached
key-value data on errors in the chain using `ExtractAuxiliary()` or on the individual error with `GetAuxiliary()`

### Benign Errors

Benign errors are errors that are mainly used to indicate a certain condition, rather than something going wrong in the 
system. An example of a benign error would be an API that returns `sql.ErrNoRows` when requesting a specific
resource. Depending on whether the resource is expected to exist or not, this may not actually be an error. 

Some clients may be calling the API to just check the existence of the resource. Nonetheless, this "error" would flood 
the logs at `ERROR` level and may disrupt error tracking tools such as [sentry](https://sentry.io/welcome/). 
The server must still return the error so that it reaches the client, however on the server, it is not seen as genuine 
error and does not need to be logged as such. With `simplerr`, it is possible to mark an error as `benign`, which allows logging middleware to detect and log
the error at a less severe level such as `INFO`.

Errors can be marked benign by either using the `Benign()` or `BenignReason()` mutators. The latter also attaches a 
reason why the error was marked benign. To detect benign errors, use the `IsBenign()` function which looks for any
benign errors in the chain of errors.

### Silent Errors

Similar to benign errors, an error can be marked as silent using the `Silent()` mutator to indicate to logging middleware to not
log this error at all. This is useful in situations where a very high amount of benign errors are flooding the logs.
To detect silent errors, use the `IsSilent()` function which looks for any silent errors in the chain of errors.

### Changing Error Formatting

The default formatting of the error string can be changed by modifying the `simplerr.Formatter` variable.
For example, to use a new line to separate the message and the wrapped error you can do:

```go
simplerr.Formatter = func(e *simplerr.SimpleError) string {
    parent := e.Unwrap()
    if parent == nil {
        return e.GetMessage()
    }
    return strings.Join([]string{e.GetMessage(), parent.Error()}, "\n")
}
```

## HTTP Status Codes

HTTP status codes can be set automatically by using the [ecosystem/http](https://github.com/lobocv/simplerr/tree/master/ecosystem/http)
package to translate `simplerr` error codes to HTTP status codes. 

```c
func (s *Server) GetUser(resp http.ResponseWriter, req *http.Request) {
	
    // extract userName from request...
	
    err := s.db.GetUser(userName)
	if err != nil {
	    // If err is a SimplError with code NotFound, the HTTP status will be set to 404
	    // If no mapping is found, the status is set to 500
	    simplehttp.SetStatus(resp, err)
	    return
    }

    resp.WriteHeader(http.StatusCreated)
}
```

Simplerr does not provide a 1:1 mapping of all HTTP status because there are too many obscure and under-utilized HTTP codes
that would complicate and bloat the library. Most of the prevalent HTTP status codes have representation in `simplerr`.
Additional translations can be added by registering a mapping:

```go
func main() {
    m := simplehttp.DefaultMapping()
    m[simplerr.CodeCanceled] = http.StatusRequestTimeout
    simplehttp.SetMapping(m)
    // ...
}
```

## GRPC Status Codes

Since GRPC functions return an error, it is even convenient to integrate error code translation using an interceptor (middleware).
The package [ecosystem/grpc](https://github.com/lobocv/simplerr/tree/master/ecosystem/http) defines an interceptor
that detects if the returned error is a `SimpleError` and then translates the error code into a GRPC status code. A mapping
for several codes is provided using the `DefaultMapping()` function. This can be changed by providing an alternative mapping
when creating the interceptor:

```go
func main() {
    // Get the default mapping provided by simplerr
    m := simplerr.DefaultMapping()
    // Add another mapping from simplerr code to GRPC code
    m[simplerr.CodeMalformedRequest] = codes.InvalidArgument
    // Create the interceptor by providing the mapping
    interceptor := simplerr.TranslateErrorCode(m)
}
```

# Contributing

Contributions and pull requests to `simplerr` are welcome but must align with the goals of the package:
- Keep it simple
- Features should have reasonable defaults but provide flexibility with optional configuration
- Keep dependencies to a minimum

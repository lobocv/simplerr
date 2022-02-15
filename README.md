[![GoReportCard](https://goreportcard.com/badge/github.com/lobocv/simplerr)](https://goreportcard.com/badge/github.com/lobocv/simplerr)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-98%25-brightgreen.svg?longCache=true&style=flat)</a>


# Simplerr

Simplerr provides a simpler and more powerful Go error handling experience by providing an alternative `error` 
implementation, the `SimpleError`. Simplerr was designed to be convenient, un-opinionated and highly customizable (if needed).
One of the main goals of Simplerr is to reduce boilerplate and make error handling and debugging easier.

# Features

The `SimpleError` allows you to easily:

- Apply an error code to any error. Choose from a list of standard codes or register your own.
- Register `func(err) *SimpleError` conversion functions to more quickly convert to `SimpleErrors` using `Convert()`.
- Automatically translate `simplerr` (including custom codes) error codes to other standardized codes such as `HTTP/gRPC`.
- Attach key-value pairs to errors to be used with structured loggers.
- Capture stack traces at the point the error is raised.
- Mark errors as `silent` so they can be skipped by logging middleware.
- Mark errors as `benign` so they can be logged less severely by logging middleware.

A complete list of standard error codes can be found [here](https://github.com/lobocv/simplerr/blob/master/codes.go).

### Basic usage

#### Creating errors
Errors can be created with `New(format string, args... interface{})`, which works similar to `fmt.Errorf` but instead
returns a `*SimplerError`. You can then chain mutations onto the error to add additional information.

```go
userID := 123
companyID := 456
err := simplerr.New("user %d does not exist in company %d", userID, companyID).
	Code(CodeNotFound).
	Aux("user_id", userID, "company_id", companyID).
```

In the above example, a new error is created and set to error code `CodeNotFound`. We have also attached auxiliary
key-value pair information to the error that we can extract later on when we decide to handle or log the error.

Errors can also be wrapped with the `Wrap(err error)` and `Wrapf(err error, format string, args... []interface{})` functions:

```go
userID := 123
err := db.GetUser(123)
if err != nil {
    serr = simplerr.Wrapf(err, "failed to get user with id = %d", userID).Aux("user_id", userID)
    if errors.Is(err, sql.ErrNoRows) {
        serr.Code(CodeNotFound)   
    }
    return serr
}
```

#### Automatic error conversion:

The above example where we manually check for `sql.ErrNoRows` can be cleaned up further by globally registering an error 
conversion function:

```go

func main() {
    // Register a conversion function for sql.ErrNoRows to be converted to CodeNotFound
    r := GetRegistry()
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
    user, err := db.GetUser(123)
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

### Detecting errors

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

### Handling and Logging SimpleErrors

`SimpleErrors` were designed to be handled. The [ecosystem](https://github.com/lobocv/simplerr/tree/master/ecosystem)
package provides packages to assist with error handling for different applications. Designing your own handlers is as 
simple as detecting the `SimpleError` and reacting to it's attributes.

#### HTTP Status Codes

HTTP status codes can be set automatically by using the [ecosystem/http](https://github.com/lobocv/simplerr/tree/master/ecosystem/http)
package to translate `simplerr` error codes to HTTP status codes. 

```go

func (s *Server) GetUser(resp http.ResponseWriter, req *http.Request) {
	
    // extract userID from request...
	
    err := db.GetUser(userID)
	if err != nil {
	    // If err is a SimplError with code NotFound, the HTTP status will be set to 404
	    // If no mapping is found, the status is set to 500
	    simplehttp.SetStatus(resp, err)
	    return
    }

    resp.WriteHeader(200)
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

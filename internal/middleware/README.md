# Middleware

Middleware are responsible for wrapping http requests or
responses.

## Middleware Definition

Middleware are defined in the `internal/middleware` directory. They
are defined according to the following guidelines:

* named for the action they perform
* are grouped as required into the same file
* are defined as a function that returns a `func(http.Handler) http.Handler`

## Example

```go
// internal/middleware/account_required.go

func AccountRequired(next http.Handler) http.Handler {
  // ...
}

```

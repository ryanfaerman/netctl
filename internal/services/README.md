# Services

Services have the responsibility of managing the business logic of
the application. They provide a layer of abstraction between the
handlers and the models, or other components.

Their full responsibilities are to:

* Enforce business rules
* Perform data validation
* Return errors to the handler
* Return data to the handler
* Trigger events (server-sent or otherwise)
* Send emails or otherwise interact with external services
* Run background tasks

## Service Definition

Services are defined in the `internal/services` directory. They should
have the following structure:

```go
// internal/services/some_service.go
type account struct {
  internalState bool
}

var Account account

func (account) Create() error {
  return nil
}

```

Notice that the service is defined as an exported variable of an
unexported type. This is since we don't really want to expose internal
details to consumers of the service and since these services are moderately
singleton in nature, they really shouldn't store much internal state.

The method names should be descriptive of the action being performed
without a stutter. Use `Create`, not `CreateAccount` nor `AccountCreate`.
Consider how consumers are using the service. For an account service,
it would be written `Account.Create`.

# Handlers

Handlers have the reponsibility of handling requests and
returning responses. They are the entry point of the
application by users.

Their full responsibilities are to:

* Extract user input from the request
* Enforce authentication requirements
* Ensure request security
* Display errors to the user
* Return HTML or JSON responses

## Handler Definition

A handler is defined in the `internal/handlers` directory. It
should be defined according to the following guidelines:

* Named in the singular form, e.g. `Account`, `Net`
* Defines a struct with exported methods
* Registers itself with into `global.handlers` on `init`
* Has a `Routes` method that defines its routes

## Example

```go
// internal/handlers/account.go

type Account struct{}

func init() {
  global.handlers = append(global.handlers, Account{})
}

func (h Account) Routes(r chi.Router)  {
  r.Get(named.Route("account-show", "/account/{account_id}"), h.Show)
}

func (h Account) Show(w http.ResponseWriter, r *http.Request) {
  // ...
}
```

Notice that we use `func(h Account)`, despite the fact that `Account`
starts with an `A`. The reason is because we're a handler.

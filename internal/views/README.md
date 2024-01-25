# Views and View Models

A view is responsible for generating the HTML to be sent to
a client by a handler or service.

It should also, optionally, define a View Model for grouping
views together with shared state.

Views are written using `templ`.

## Example

```go

type Account struct {
  Account *models.Account
  Accounts []*models.Account
}

templ (v Account) List() string {
  <h1>Account List</h1>
  <ul>
    for _, account := range v.Accounts {
      @v.AccountRow(account)
    }
  </ul>
}

templ (v Account) AccountRow(account *models.Account) string {
  <li>{ account.Name }</li>
}

```

# Models

Models have the responsibility of representing how data is structured
within the application and defining validation rules.

Their full responsibilities are to:

* Define data structures and their relationships
* Define validation rules
* Define encoding and decoding rules (json, xml, etc.)
* Provide finder methods to lookup/query data

## Model Definition

A model is defined in the `internal/models` directory. It should be defined
according to the following guidelines:

* Named in the singular form, e.g. `Account`, `Net`
* Exist in its own file, e.g. `internal/models/account.go`
* Contain a struct definition with exported fields
* Define a `New{Model}` constructor function
* Children models can either be defined within the same file or in its own file `internal/models/account.preference.go`
  * When using a subfile, the model would be `AccountPreference`
  * Slices of a model should be defined within the model file
* An ID field should always be capitalized
* Model methods may load additional data (and queries) as required, but should not
be responsible for saving data
* Is responsible for defining validation rules for data integrity only

### A note about finders

Finders should be descriptive. They should be named according to the arguments and what
they're finding.

If you're querying for an account by id, the method would be invoked in way that
indicates that to consumers: `models.Account.FindByID(3)`. This approach makes
the consuming code self-documenting in some sense.

If the finder returns more than one result, it should be named in the plural
form, otherwise use the singular form.

## Example

```go
// internal/models/account.go
type Account struct {
  ID        int
  Name      string `validate:"required" json:"name"`
  CreatedAt time.Time
  Active    bool
  Preferences AccountPreferences
}

func NewAccount() *Account {
  // ...
}

func FindAccountByID(id int) (*Account, error) {
  // ...
}

func FindInactiveAccounts() ([]*Account, error) {
  // ...
}

func (m *Account) Age() (time.Duration, error) {
  // ...
}

```

```go
// internal/models/account.preference.go
type AccountPreference struct {
  // ...
}

func NewAccountPreference() *AccountPreference {
  // ...
}

type AccountPreferences []*AccountPreference
```

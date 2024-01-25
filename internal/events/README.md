# Events

 Events are signals that something has happened. It contains the
 pertinent information about that signal, but should not contain
 any functionality.

Their full responsibilities are to:

* Define the event structure
* Contain data about the event

## Event Definition

An event is defined in the `internal/events` directory. It should be defined
according to the following guidelines:

* The filename is the category or prefix of the event
* The event is named with a category or prefix
* It should contain a string ID (generally will be a ULID)
* It should only have exported fields
* It only uses primitive types (string, int, etc.)
* It does not use any interface types

## Example

```go
// internal/events/account.go

type (
  AccountCreated struct{
    ID   string
    Name string
  }

  AccountBanned struct {
    ID string
  }

  AccountPreferenceUpdated struct {
    ID    string
    Value string

    ByAccount string
  }
)

func init() {
  gob.Register(AccountCreated{})
  gob.Register(AccountBanned{})
  gob.Register(AccountPreferenceUpdated{})
}
```

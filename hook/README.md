# Hooks

## Overview
_Hooks_ provides a simple, **type-safe** hook system to enable easier
modularization of your Go code. A _hook_ allows various parts of your codebase
to tap into events and operations happening elsewhere which prevents direct
coupling between the producer and the consumers/listeners.

For example, a _user_ package/module in your code may dispatch a _hook_ when a
user is created, allowing your _notification_ package to send the user an
email, and a _history_ package to record the activity without the _user_ module
having to call these components directly. A hook can also be used to allow
other modules to alter and extend data before it is processed.

Hooks can be very beneficial especially in a monolithic application both for
overall organization as well as in preparation for the splitting of modules
into separate synchronous or asynchronous services.

This is based on https://github.com/mikestefanello/hooks and extended heartily.

## Usage


1) Start by declaring a new hook which requires specifying the _type_ of data that it will dispatch as well as a name. This can be done in a number of different way such as a global variable or exported field on a _struct_:

```go
package user

type User struct {
    ID int
    Name string
    Email string
    Password string
}

var HookUserInsert = hooks.NewHook[User]("user.insert")
```

2) Register for a hook:

```go
package greeter

func init() {
    user.HookUserInsert.Register(func(e hooks.Event[user.User]) {
        sendEmail(e.Msg.Email)
    })
}
```

3) Dispatch the data to the hook _Registrants_:

```go
func (u *User) Insert() {
    db.Insert("INSERT INTO users ...")

    HookUserInsert.Dispatch(context.Background(), &u)
}
```

### Managing concurrency

Hooks are dispatched, by default, concurrently with a concurrency of 1. That
means you can raise the limit and increase the concurrency. You define this
when the hook is created.

Using the `HookUserInsert` example above:

```go
var HookUserInsert = hooks.NewHook[User]("user.insert").WithLimit(10)
```

With the limit raised to 10, multiple registrants can concurrently execute.
Bounded by a rate of 10 at once. If a registrant finishes before another, it
frees capacity and another may start.

This is an important distinction. The work doesn't work in batches. If the
concurrency is set to N, we don't wait until all N are complete before starting
more work. As each item completes, it frees a token for another to work.

The dispatcher will block until all registrants finish.

The `context.Context` argument let's the dispatcher cancel the work or setup a
deadline. This way, should the caller no longer need the work, we're not
burning cycles.

## Using hooks as a message bus

A registrant can `Unregister` which means our basic hook/notification approach
can be used as a sort of simple message bus. Let's walk through this with an example.

1) Start by declaring a new hook which requires specifying the _type_ of data that it will dispatch as well as a name. This is the same as before.
2) Register for a hook and *Unregister* on some condition:

```go
package greeter

func init() {
    user.HookUserInsert.Register(func(e hooks.Event[user.User]) {
      if someCondition() {
        e.Unregister()
      }
      sendEmail(e.Msg.Email)
    })
}
```

3) Dispatch the data to the hook _Registrants_:

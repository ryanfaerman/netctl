package hook_test

import (
	"context"
	"fmt"

	"github.com/ryanfaerman/netctl/hook"
)

func ExampleHook_Register() {
	type Message struct {
		Greeting string
	}
	example := hook.New[Message]("test.hook")

	example.Register(func(e hook.Event[Message]) {
		fmt.Println(e.Payload.Greeting)
	})

	example.Dispatch(context.Background(), Message{Greeting: "Hello There!"})
	// Output: Hello There!

}

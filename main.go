package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/events"
)

// Define a sample struct
type YourStruct struct {
	ID       string `json:"id"`
	Callsign string `json:"callsign"`
	Name     string `json:"name"`
	Location string `json:"location"`
	// ... other fields
}

var registry = make(map[string]reflect.Type)

func register(e interface{}) {
	name := fmt.Sprintf("%T", e)
	fmt.Println("registering", name)
	registry[name] = reflect.TypeOf(e)
}

func get(name string) interface{} {
	return reflect.New(registry[name]).Interface()
}

func get2(name string) interface{} {
	instance := reflect.New(registry[name]).Elem().Interface()
	return reflect.Indirect(reflect.ValueOf(instance)).Interface()
}

func get3[K any](name string) K {
	instance := reflect.New(registry[name]).Elem().Interface()
	return reflect.Indirect(reflect.ValueOf(instance)).Interface().(K)
}

var handlers = make(map[string]func() any)

func handle[K any]() {
	n := new(K)
	spew.Dump(n)
	name := fmt.Sprintf("%T", *new(K))
	fmt.Println(name)
	handlers[name] = func() any {
		return new(K)
	}
}

func decode(kind string, data []byte) any {
	k := handlers[kind]()
	json.Unmarshal(data, k)
	return k
}

func main() {
	// register(events.NetCheckinHeard{})
	// register(YourStruct{})
	//
	// k := get2("events.NetCheckinHeard")
	// spew.Dump(k)
	//
	// data := `{"id":"01HNDEHNP9BJSGYS5369TPQV04","callsign":"KQ4JXI","name":"","location":"","kind":"Routine","traffic":0}`
	// json.Unmarshal([]byte(data), &k)
	// spew.Dump(k)
	//
	// name := fmt.Sprintf("%T", k)
	// fmt.Println(name)
	// handle[events.NetCheckinHeard]()
	//
	// data := `{"id":"01HNDEHNP9BJSGYS5369TPQV04","callsign":"KQ4JXI","name":"","location":"","kind":"Routine","traffic":0}`
	// fmt.Println("whoop")
	// k := decode("events.NetCheckinHeard", []byte(data))
	// spew.Dump(k)

	data := `{"id":"01HNDEHNP9BJSGYS5369TPQV04","callsign":"KQ4JXI","name":"","location":"","kind":"Routine","traffic":0}`
	k, err := events.Decode("events.NetCheckinHeard", []byte(data))
	if err != nil {
		log.Fatal(err)
	}
	switch v := k.(type) {
	case *events.NetCheckinHeard:
		fmt.Println(v.Callsign)
		spew.Dump(v)
	}
}

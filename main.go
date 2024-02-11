package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/models"
)

func main() {
	spew.Dump(
		uint(models.PermissionEdit),
		uint(models.PermissionRunNet),
		uint(models.PermissionDingle),
	)

	fmt.Println("custom", uint(models.PermissionEdit|models.PermissionDingle))
}

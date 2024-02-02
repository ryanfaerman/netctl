package main

import (
	"fmt"

	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
)

type dingus struct{}

func (d *dingus) Verbs() []string {
	return []string{"create"}
}

func main() {
	can("create", &models.NetCheckin{})
	can("create", &dingus{})
	can("delete", &dingus{})
}

func can(action string, m any) {
	accounts := []*models.Account{
		models.AccountAnonymous,
		{ID: 17, Name: "fred"},
		{ID: 1, Name: "tom"},
	}
	for i, a := range accounts {
		fmt.Printf("Can '%s' be done on '%T' by '%s'?  ", action, m, a.Name)

		if err := services.Authorization.Can(a, action, m); err != nil {
			fmt.Printf("[NO]; %s\n", err.Error())
		} else {
			fmt.Print("[YES]\n")
		}
		if i > 0 {
			fmt.Println("---")
		}
	}
}

package services

import (
	"errors"
	"fmt"
	"slices"

	"github.com/ryanfaerman/netctl/internal/models"
)

// Canner allows a resource to define if an account can perform a given action.
type Canner interface {
	Can(account *models.Account, action string) error
}

// Verber is used to allow a resource to define the actions it supports.
type Verber interface {
	Verbs() []string
}

var (
	ErrNotAuthorized = errors.New("not authorized")
	ErrInvalidAction = errors.New("invalid action")
)

// authz is our service for authorization.
type authz struct {
	Policy Policy
}

// Define our global authorization service.
var Authorization = authz{
	Policy: DefaultPolicy,
}

// Policy is a map of actions to functions that
// return an error if the action cannot be performed for the given account.
type Policy map[string]func(*models.Account) error

// Verbs implements the Verber interface.
func (p Policy) Verbs() []string {
	var verbs []string
	for verb := range p {
		verbs = append(verbs, verb)
	}
	return verbs
}

// Can implements the Canner interface.
func (p Policy) Can(account *models.Account, action string) error {
	if fn, ok := p[action]; ok {
		return fn(account)
	}
	return nil
}

var DefaultPolicy = Policy{
	"view-metrics": func(account *models.Account) error {
		if account.IsAnonymous() {
			return ErrNotAuthorized
		}
		return nil
	},
	"manage-account": func(account *models.Account) error {
		if account.IsAnonymous() {
			return ErrNotAuthorized
		}
		return nil
	},
}

// Can checks if the given account can perform the given action on _all_
// of the provided resources.
//
// A resource is anything that implements the Canner interface, if no
// resources are provided, the DefaultPolicy is used.
//
// If the account is nil, it is assumed to be an anonymous account.
// If no resources are provided, the DefaultPolicy is used.
//
// A wrapped error is returned with with ErrNotAuthorized as
// the first and the specific error as the second.
//
// If the action is not valid for the resource, ErrInvalidAction
// is returned. Where validity is defined by the resource
// implementing the optional Verber interface.
func (a *authz) Can(account *models.Account, action string, resources ...any) error {
	l := global.log.With("action", action, "service", "authorization")
	if account == nil {
		account = models.AccountAnonymous
	}
	l = l.With("account", account.ID, "anon", account.IsAnonymous())

	if len(resources) == 0 {
		l.Debug("using internal policy")
		resources = append(resources, a.Policy)
	}

	for _, resource := range resources {
		l = l.With("resource", fmt.Sprintf("%T", resource))
		if r, ok := resource.(Verber); ok {
			if !slices.Contains(r.Verbs(), action) {
				l.Debug("invalid action", "err", ErrInvalidAction.Error())
				return ErrInvalidAction
			}
		}

		if r, ok := resource.(Canner); ok {
			if err := r.Can(account, action); err != nil {
				err = errors.Join(ErrNotAuthorized, err)
				l.Debug("not authorized", "err", err.Error())
				return err
			}
		}
	}
	l.Debug("authorized")
	return nil
}

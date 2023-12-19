package named

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type route struct {
	path   string
	params []string
}

func (r route) build(params ...string) (string, error) {
	if len(params) != len(r.params) {
		return "", errors.New("parameter length mismatch")
	}

	s := r.path
	for i, v := range params {
		s = strings.Replace(s, r.params[i], v, 1)
	}

	return s, nil
}

func (r route) String() string {
	return r.path
}

type Store struct {
	routes map[string]route
	l      sync.RWMutex
	re     *regexp.Regexp
}

func NewStore() *Store {
	return &Store{
		routes: make(map[string]route),
		re:     regexp.MustCompile(`{\w+|\*}`),
	}
}

func (s *Store) SetParamPattern(p string) error {
	re, err := regexp.Compile(p)
	if err != nil {
		return err
	}

	s.re = re
	return nil
}

func (s *Store) Add(name string, path string, params ...string) (string, error) {
	s.l.Lock()
	defer s.l.Unlock()

	if _, ok := s.routes[name]; ok {
		return "", errors.New(fmt.Sprintf("routed named '%s' already exists!", name))
	}

	s.routes[name] = route{path: path, params: params}
	return path, nil
}

func (s *Store) MustAdd(name string, path string, params ...string) string {
	p, err := s.Add(name, path, params...)
	if err != nil {
		panic(err.Error())
	}
	return p
}

func (s *Store) Resolve(name string, params ...string) (string, error) {
	s.l.RLock()
	defer s.l.RUnlock()

	r, ok := s.routes[name]
	if !ok {
		return "", errors.New("unknown route")
	}

	return r.build(params...)
}

func (s *Store) MustResolve(name string, params ...string) string {
	p, err := s.Resolve(name, params...)
	if err != nil {
		panic(err.Error())
	}
	return p
}

func (s *Store) Lookup(path string) string {
	// Lookup the path in the store and return the name of the route
	s.l.RLock()
	defer s.l.RUnlock()

	for name, r := range s.routes {
		if r.path == path {
			return name
		}
	}

	return ""
}

var Default = NewStore()

func Route(name, path string) string {
	matches := Default.re.FindAllStringSubmatch(path, -1)
	params := []string{}
	if matches != nil {
		for _, param := range matches {
			params = append(params, param[0])
		}
	}
	return Default.MustAdd(name, path, params...)
}

func URLFor(name string, params ...string) string {
	return Default.MustResolve(name, params...)
}

func Lookup(path string) string {
	return Default.Lookup(path)
}

// view_profile: /profile/{name}

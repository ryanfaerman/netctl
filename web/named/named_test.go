package named

import "testing"

func TestUrlBuild(t *testing.T) {
	examples := []struct {
		route  route
		params []string
		out    string
		err    error
	}{
		{
			route:  route{path: "/foo"},
			params: []string{},
			out:    "/foo",
		},
		{
			route:  route{path: "/foo/{p1}/{p2}", params: []string{"{p1}", "{p2}"}},
			params: []string{"bar", "baz"},
			out:    "/foo/bar/baz",
		},
	}

	for _, example := range examples {
		t.Run(example.route.String(), func(t *testing.T) {
			actual, err := example.route.build(example.params...)
			if err != example.err {
				t.Errorf("unexpected error; expected '%v', got '%v'", example.err, err)
				return
			}
			if actual != example.out {
				t.Errorf("incorrect result; expected: '%s', go '%s'", example.out, actual)
			}
		})

	}
}

func TestAdd(t *testing.T) {
	examples := []struct {
		name string
		path string
		out  string
	}{
		{
			name: "foo",
			path: "/foo",
			out:  "/foo",
		},
	}

}

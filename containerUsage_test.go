package main

import "testing"

func Test_NextGitlabPage(t *testing.T) {
	for name, tc := range map[string]struct {
		header string
		exp    string
	}{
		"upstream-exampe": {
			header: `<https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=1&per_page=3>; rel="prev", <https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=3&per_page=3>; rel="next", <https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=1&per_page=3>; rel="first", <https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=3&per_page=3>; rel="last"`,

			exp: "https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=3&per_page=3",
		},
		"no more pages": {
			header: ``,
			exp:    "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			name, tc := name, tc
			t.Parallel()
			got := nextGitlabPage(tc.header)
			if got != tc.exp {
				t.Errorf("test %s\ngot %q\nwant %q", name, got, tc.exp)
			}
		})
	}
}

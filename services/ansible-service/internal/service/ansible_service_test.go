package service

import "testing"

func TestParseGitRepo(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://forgejo.client.interne/ops/playbooks.git", "ops/playbooks"},
		{"https://codeberg.org/kura/ansible-playbooks", "kura/ansible-playbooks"},
		{"git@forgejo.client.interne:ops/playbooks.git", "ops/playbooks"},
		{"https://host/group/sub/repo.git", "sub/repo"},
		{"", ""},
		{"https://host/seulement", ""},
	}
	for _, c := range cases {
		if got := ParseGitRepo(c.in); got != c.want {
			t.Errorf("ParseGitRepo(%q) = %q, attendu %q", c.in, got, c.want)
		}
	}
}

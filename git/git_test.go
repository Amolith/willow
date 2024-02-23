package git

import (
	"testing"
)

func TestStringifyRepo(t *testing.T) {
	wantGitHub := "data/github.com/owner/repo"
	wantSourceHut := "data/git.sr.ht/~owner/repo"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "GitHubHTTP",
			input: "http://github.com/owner/repo",
			want:  wantGitHub,
		},
		{
			name:  "GitHubHTTPS",
			input: "https://github.com/owner/repo",
			want:  wantGitHub,
		},
		{
			name:  "GitHubSSH",
			input: "git@github.com:owner/repo",
			want:  wantGitHub,
		},
		{
			name:  "SourceHutHTTP",
			input: "http://git.sr.ht/~owner/repo",
			want:  wantSourceHut,
		},
		{
			name:  "SourceHutHTTPS",
			input: "https://git.sr.ht/~owner/repo",
			want:  wantSourceHut,
		},
		{
			name:  "SourceHutSSH",
			input: "git@git.sr.ht:~owner/repo",
			want:  wantSourceHut,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := stringifyRepo(test.input)
			if err != nil {
				t.Errorf("stringifyRepo(%s) returned error: %v", test.input, err)
			}
			if got != test.want {
				t.Errorf("stringifyRepo(%s) = %s, want %s", test.input, got, test.want)
			}
		})
	}
}

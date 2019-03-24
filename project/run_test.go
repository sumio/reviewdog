package project

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/reviewdog/reviewdog"
)

type fakeDiffService struct {
	reviewdog.DiffService
	FakeDiff func() ([]byte, error)
}

func (f *fakeDiffService) Diff(_ context.Context) ([]byte, error) {
	return f.FakeDiff()
}

func (f *fakeDiffService) Strip() int {
	return 0
}

type fakeCommentService struct {
	reviewdog.CommentService
	FakePost func(*reviewdog.Comment) error
}

func (f *fakeCommentService) Post(_ context.Context, c *reviewdog.Comment) error {
	return f.FakePost(c)
}

func TestRun(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		conf := &Config{}
		if err := Run(ctx, conf, nil, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("erorformat error", func(t *testing.T) {
		conf := &Config{
			Runner: map[string]*Runner{
				"test": {},
			},
		}
		if err := Run(ctx, conf, nil, nil); err == nil {
			t.Error("want error, got nil")
		} else {
			t.Log(err)
		}
	})

	t.Run("diff error", func(t *testing.T) {
		ds := &fakeDiffService{
			FakeDiff: func() ([]byte, error) {
				return nil, errors.New("err!")
			},
		}
		conf := &Config{
			Runner: map[string]*Runner{
				"test": {
					Cmd:         "not found",
					Errorformat: []string{`%f:%l:%c:%m`},
				},
			},
		}
		if err := Run(ctx, conf, nil, ds); err == nil {
			t.Error("want error, got nil")
		} else {
			t.Log(err)
		}
	})

	t.Run("no cmd error (not for reviewdog to exit with error)", func(t *testing.T) {
		ds := &fakeDiffService{
			FakeDiff: func() ([]byte, error) {
				return []byte(""), nil
			},
		}
		cs := &fakeCommentService{
			FakePost: func(c *reviewdog.Comment) error {
				return nil
			},
		}
		conf := &Config{
			Runner: map[string]*Runner{
				"test": {
					Cmd:         "not found",
					Errorformat: []string{`%f:%l:%c:%m`},
				},
			},
		}
		if err := Run(ctx, conf, cs, ds); err != nil {
			t.Error(err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ds := &fakeDiffService{
			FakeDiff: func() ([]byte, error) {
				return []byte(""), nil
			},
		}
		cs := &fakeCommentService{
			FakePost: func(c *reviewdog.Comment) error {
				return nil
			},
		}
		conf := &Config{
			Runner: map[string]*Runner{
				"test": {
					Cmd:         "echo 'hi'",
					Errorformat: []string{`%f:%l:%c:%m`},
				},
			},
		}
		if err := Run(ctx, conf, cs, ds); err != nil {
			t.Error(err)
		}
	})

}

func TestFilteredEnviron(t *testing.T) {
	names := [...]string{
		"REVIEWDOG_GITHUB_API_TOKEN",
		"REVIEWDOG_GITLAB_API_TOKEN",
		"REVIEWDOG_TOKEN",
	}

	for _, name := range names {
		defer func(name, value string) {
			os.Setenv(name, value)
		}(name, os.Getenv(name))
		os.Setenv(name, "value")
	}

	filtered := filteredEnviron()
	if len(filtered) != len(os.Environ())-len(names) {
		t.Errorf("len(filtered) != len(os.Environ())-%d, %v != %v-%d", len(names), len(filtered), len(os.Environ()), len(names))
	}

	for _, kv := range filtered {
		for _, name := range names {
			if strings.HasPrefix(kv, name) && kv != name+"=" {
				t.Errorf("filtered: %v, want %v=", kv, name)
			}
		}
	}

	for _, kv := range os.Environ() {
		for _, name := range names {
			if strings.HasPrefix(kv, name) && kv != name+"=value" {
				t.Errorf("envs: %v, want %v=value", kv, name)
			}
		}
	}
}

// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package config_test

import (
	"testing"
	"testing/fstest"

	"github.com/nil-go/konf"
	"github.com/nil-go/konf/provider/env"

	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/internal/assert"
)

func TestNew(t *testing.T) {
	testcases := []struct {
		description string
		opts        []config.Option
		env         map[string]string
		key         string
		explanation string
		err         string
	}{
		{
			description: "config file",
			opts:        []config.Option{config.WithFile("testdata/config.yaml")},
			key:         "nilgo.source",
			explanation: `nilgo.source has value[file] that is loaded by loader[fs:///testdata/config.yaml].

`,
		},
		{
			description: "multiple config files",
			opts:        []config.Option{config.WithFile("testdata/config.yaml", "testdata/staging.yaml")},
			key:         "nilgo.stage",
			explanation: `nilgo.stage has value[staging] that is loaded by loader[fs:///testdata/staging.yaml].
Here are other value(loader)s:
  - dev(fs:///testdata/config.yaml)

`,
		},
		{
			description: "with environment variables",
			opts:        []config.Option{config.WithFile("testdata/config.yaml")},
			env:         map[string]string{"NILGO_SOURCE": "env"},
			key:         "nilgo.source",
			explanation: `nilgo.source has value[env] that is loaded by loader[env:*].
Here are other value(loader)s:
  - file(fs:///testdata/config.yaml)

`,
		},
		{
			description: "config file path in environment variable",
			env:         map[string]string{"CONFIG_FILE": "testdata/config.yaml"},
			key:         "nilgo.source",
			explanation: `nilgo.source has value[file] that is loaded by loader[fs:///testdata/config.yaml].

`,
		},
		{
			description: "default config file not found",
			key:         "nilgo.source",
			explanation: `nilgo.source has no configuration.

`,
		},
		{
			description: "with fs",
			opts:        []config.Option{config.WithFS(fstest.MapFS{"config/config.yaml": {Data: []byte("nilgo:\n  source: fs")}})},
			key:         "nilgo.source",
			explanation: `nilgo.source has value[fs] that is loaded by loader[fs:///config/config.yaml].

`,
		},
		{
			description: "with option",
			opts: []config.Option{
				config.WithOption(konf.WithCaseSensitive()),
				config.WithFS(fstest.MapFS{"config/config.yaml": {Data: []byte("nilgo:\n  source: fs")}}),
			},
			key: "nilgo.Source",
			explanation: `nilgo.Source has no configuration.

`,
		},
		{
			description: "unmarshal error",
			opts:        []config.Option{config.WithFS(fstest.MapFS{"config/config.yaml": {Data: []byte("nilgo")}})},
			key:         "nilgo.source",
			err: "load config file config/config.yaml: load configuration: unmarshal: yaml: unmarshal errors:\n" +
				"  line 1: cannot unmarshal !!str `nilgo` into map[string]interface {}",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			for key, value := range testcase.env {
				t.Setenv(key, value)
			}

			// Reset the default config to load the new environment variables.
			cfg := konf.New()
			// Ignore error: env loader does not return error.
			_ = cfg.Load(env.New())
			konf.SetDefault(cfg)

			cfg, err := config.New(testcase.opts...)
			if testcase.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, testcase.explanation, cfg.Explain(testcase.key))
			} else {
				assert.EqualError(t, err, testcase.err)
			}
		})
	}
}

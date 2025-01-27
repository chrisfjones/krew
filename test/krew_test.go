// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package test contains integration tests for krew.
package test

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"sigs.k8s.io/krew/test/krew"
)

const (
	// validPlugin is a valid plugin with a small download size
	validPlugin = "konfig"
)

func TestKrewHelp(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	test.Krew("help").RunOrFail()
}

func TestUnknownCommand(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	if err := test.Krew("foobar").Run(); err == nil {
		t.Errorf("Expected `krew foobar` to fail")
	}
}

func TestKrewInstall(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin).RunOrFailOutput()
	test.Call(validPlugin, "--help").RunOrFail()
}

func TestKrewUninstall(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin).RunOrFailOutput()
	test.Krew("uninstall", validPlugin).RunOrFailOutput()
	if err := test.Call(validPlugin, "--help").Run(); err == nil {
		t.Errorf("Expected the plugin to be uninstalled")
	}
}

func TestKrewSearchAll(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	output := test.WithIndex().Krew("search").RunOrFailOutput()
	if plugins := lines(output); len(plugins) < 10 {
		// the first line is the header
		t.Errorf("Expected at least %d plugins", len(plugins)-1)
	}
}

func TestKrewSearchOne(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	plugins := lines(test.WithIndex().Krew("search", "krew").RunOrFailOutput())
	if len(plugins) < 2 {
		t.Errorf("Expected krew to be a valid plugin")
	}
	if !strings.HasPrefix(plugins[1], "krew ") {
		t.Errorf("The first match should be krew")
	}
}

func TestKrewInfo(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("info", validPlugin).RunOrFail()
}

func TestKrewInfoInvalidPlugin(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	plugin := "invalid-plugin"
	err := test.WithIndex().Krew("info", plugin).Run()
	if err == nil {
		t.Errorf("Expected `krew info %s` to fail", plugin)
	}
}

func TestKrewList(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list").RunOrFailOutput()
	if bytes.Contains(initialList, []byte(validPlugin)) {
		t.Errorf("%q should initially not be installed", validPlugin)
	}

	test.Krew("install", validPlugin).RunOrFail()

	eventualList := test.Krew("list").RunOrFailOutput()
	if !bytes.Contains(eventualList, []byte(validPlugin)) {
		t.Errorf("%q should eventually be installed", validPlugin)
	}
}

func TestKrewVersion(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	output := test.Krew("version").RunOrFailOutput()

	requiredSubstrings := []string{
		"IsPlugin",
		fmt.Sprintf(`BasePath\s+%s`, test.Root()),
		"ExecutedVersion",
		"GitTag",
		"GitCommit",
		`IndexURI\s+https://github.com/kubernetes-sigs/krew-index.git`,
		"IndexPath",
		"InstallPath",
		"DownloadPath",
		"BinPath",
	}

	for _, p := range requiredSubstrings {
		if regexp.MustCompile(p).FindSubmatchIndex(output) == nil {
			t.Errorf("Expected to find %q in output of `krew version`", p)
		}
	}
}

func TestKrewUpdate(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewTest(t)
	defer cleanup()

	// nb do not call WithIndex() here
	test.Krew("update").RunOrFail()
	plugins := lines(test.Krew("search").RunOrFailOutput())
	if len(plugins) < 10 {
		// the first line is the header
		t.Errorf("Less than %d plugins found, `krew update` most likely failed unless TestKrewSearchAll also failed", len(plugins)-1)
	}
}

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

func lines(in []byte) []string {
	trimmed := strings.TrimRight(string(in), " \t\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

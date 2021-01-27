/*
Copyright © 2020 Chris Winkler

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package utils

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type (
	// Config represents a configuration for Dux
	Config struct {
		Namespaces map[string]Namespace
		Managers   map[string]Manager
	}

	// Namespace represents the config for a local package list
	Namespace struct {
		ListCmd        string `mapstructure:"list-cmd"`
		RemoveCmd      string `mapstructure:"remove-cmd"`
		DefaultManager string `mapstructure:"default-manager"`
	}

	// Manager represents the config for a package source / installer
	Manager struct {
		Namespace  string
		InstallCmd string `mapstructure:"install-cmd"`
		ExpandCmd  string `mapstructure:"expand-cmd"`
	}
)

// Validate checks the config object for semantic errors
func (c *Config) Validate() error {
	if len(c.Namespaces) == 0 {
		return errors.New("No namespaces found in config")
	}
	if len(c.Managers) == 0 {
		return errors.New("No managers found in config")
	}

	for name, ns := range c.Namespaces {
		// validate each pkglist
		if err := ns.validateWith(c.Managers); err != nil {
			return fmt.Errorf("Validation error for namespace \"%s\": %w", name, err)
		}
	}

	for name, mgr := range c.Managers {
		// validate each pkgsource
		if err := mgr.validateWith(c.Namespaces); err != nil {
			return fmt.Errorf("Validation error for manager \"%s\": %w", name, err)
		}
	}

	return nil
}

func (ns *Namespace) validateWith(mgrs map[string]Manager) error {
	if len(ns.ListCmd) == 0 {
		return errors.New("missing `list-cmd`")
	}
	if len(ns.RemoveCmd) == 0 {
		return errors.New("missing `remove-cmd`")
	}
	if len(ns.DefaultManager) == 0 {
		return errors.New("missing `default-manager`")
	}

	for mgrName := range mgrs {
		if mgrName == ns.DefaultManager {
			return nil
		}
	}
	return fmt.Errorf("Default source `%s` does not exist", ns.DefaultManager)
}

func (mgr *Manager) validateWith(nss map[string]Namespace) error {
	if len(mgr.Namespace) == 0 {
		return errors.New("missing `namespace`")
	}

	for nsName := range nss {
		if nsName == mgr.Namespace {
			return nil
		}
	}
	return fmt.Errorf("namespace `%s` does not exist", mgr.Namespace)

}

// ConfigFromViper parses a config obj from Viper and validates it
func ConfigFromViper() (*Config, error) {
	cfg := &Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("Config parse error: %v", err)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("Config validation error: %v", err)
	}

	return cfg, nil
}

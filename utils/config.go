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
		PkgLists []PkgList
		Sources  []PkgSource
	}

	// PkgList represents the config for a local package list
	PkgList struct {
		Name          string
		ListCmd       string `mapstructure:"list"`
		RemoveCmd     string `mapstructure:"remove"`
		DefaultSource string `mapstructure:"default-source"`
	}

	// PkgSource represents the config for a package source
	PkgSource struct {
		Name       string
		Default    bool
		PkgList    string
		InstallCmd string `mapstructure:"install"`
	}
)

// Validate checks the config object for semantic errors
func (c *Config) Validate() error {
	if len(c.PkgLists) == 0 {
		return errors.New("No `pkglist`s found in config")
	}
	if len(c.Sources) == 0 {
		return errors.New("No `source`s found in config")
	}

	pkgListNames := make(map[string]struct{})
	for i, pkgList := range c.PkgLists {
		// check for duplicate names
		if _, ok := pkgListNames[pkgList.Name]; ok {
			return errors.New("Name collision")
		}
		pkgListNames[pkgList.Name] = struct{}{}

		// validate each pkglist
		if err := pkgList.validateWith(c.Sources); err != nil {
			return fmt.Errorf("Validation error at pkglists[%d]: %w", i, err)
		}
	}

	pkgSourceNames := make(map[string]struct{})
	pkgSourceDefault := false
	for i, pkgSource := range c.Sources {
		// check for duplicate names
		if _, ok := pkgSourceNames[pkgSource.Name]; ok {
			return errors.New("Name collision")
		}
		pkgSourceNames[pkgSource.Name] = struct{}{}

		// Make sure there isn't more than one package source tagged as "default"
		if pkgSource.Default {
			if pkgSourceDefault {
				return fmt.Errorf("More than one pkg source listed as default: [%d]", i)
			}
			pkgSourceDefault = true
		}

		// validate each pkgsource
		if err := pkgSource.validateWith(c.PkgLists); err != nil {
			return fmt.Errorf("Validation error at sources[%d]: %w", i, err)
		}
	}

	return nil
}

func (list *PkgList) validateWith(sources []PkgSource) error {
	if len(list.Name) == 0 {
		return errors.New("missing `name`")
	}
	if len(list.ListCmd) == 0 {
		return errors.New("missing `list` command")
	}
	if len(list.RemoveCmd) == 0 {
		return errors.New("missing `remove` command")
	}
	if len(list.DefaultSource) == 0 {
		return errors.New("missing a default source `default-source`")
	}

	for _, s := range sources {
		if s.Name == list.DefaultSource {
			return nil
		}
	}
	return fmt.Errorf("Default source `%v` does not exist", list.DefaultSource)
}

func (source *PkgSource) validateWith(lists []PkgList) error {
	if len(source.Name) == 0 {
		return errors.New("missing `name`")
	}
	if len(source.PkgList) == 0 {
		return errors.New("missing `pkglist`")
	}

	for _, s := range lists {
		if s.Name == source.PkgList {
			return nil
		}
	}
	return fmt.Errorf("pkglist `%v` does not exist", source.PkgList)

}

// GetSource gets the package source with the specified name
func (c *Config) GetSource(name string) (PkgSource, bool) {
	for _, p := range c.Sources {
		if p.Name == name {
			return p, true
		}
	}
	return PkgSource{}, false
}

// GetList gets the package list with the specified name
func (c *Config) GetList(name string) (PkgList, bool) {
	for _, p := range c.PkgLists {
		if p.Name == name {
			return p, true
		}
	}
	return PkgList{}, false
}

// GetDefaultSource gets the PkgSource that doesn't require an explicit tag in the target files
func (c *Config) GetDefaultSource() PkgSource {
	for _, src := range c.Sources {
		if src.Default {
			return src
		}
	}
	return PkgSource{}
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

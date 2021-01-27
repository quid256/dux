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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GetInstalledPackages runs the appropriate shell commands to retrieve all of the packages installed in a given namespace
func GetInstalledPackages(ns Namespace) (map[string]struct{}, error) {

	out, err := exec.Command("bash", "-c", ns.ListCmd).Output()
	if err != nil {
		return nil, fmt.Errorf("Error executing list-cmd: %w", err)
	}

	s := strings.Split(string(out), "\n")
	pkgs := make(map[string]struct{}, len(s))

	for _, l := range strings.Split(string(out), "\n") {
		if len(l) == 0 {
			continue
		}
		pkgs[l] = struct{}{}
	}

	return pkgs, nil
}

// GetTargetPackages gets the target packages from the target directory with the config
func GetTargetPackages(cfg *Config, targetDir string) (map[string](map[string]string), error) {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Missing target directory: %s. Perhaps run `dux generate`", targetDir)
	}

	mgrToTargets := make(map[string][]string)

	commentRe := regexp.MustCompile(`#[^\n]*(\n|\z)`)
	identRe := regexp.MustCompile(`[^ \r\n\t"]+`)

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		contentsNoComments := commentRe.ReplaceAllString(string(contents), "")

		curSrc := ""
		tempSrc := ""

		for _, ident := range identRe.FindAllString(contentsNoComments, -1) {
			if ident[0] == '[' && ident[len(ident)-1] == ']' {
				curSrc = ident[1 : len(ident)-1]
			} else if ident[0] == '(' && ident[len(ident)-1] == ')' {
				if tempSrc != "" {
					return fmt.Errorf("Cannot have 2 adjacent temporary sources: (%s) and (%s)", tempSrc, ident[1:len(ident)-1])
				}

				tempSrc = ident[1 : len(ident)-1]
			} else {
				var source string

				if tempSrc != "" {
					source = tempSrc
					tempSrc = ""
				} else if curSrc != "" {
					source = curSrc
				} else {
					return fmt.Errorf("No source defined for package: %s", ident)
				}

				if _, ok := mgrToTargets[source]; !ok {
					mgrToTargets[source] = nil
				}

				mgrToTargets[source] = append(mgrToTargets[source], ident)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error gathering targets: %w", err)
	}

	// maps namespaces to (packageName -> manager)
	targetPkgs := make(map[string](map[string]string))

	for mgrName, targets := range mgrToTargets {
		mgr, ok := cfg.Managers[mgrName]
		if !ok {
			return nil, fmt.Errorf("No such source found: %s", mgrName)
		}

		var packageNames []string

		if mgr.ExpandCmd != "" {
			cmd := exec.Command("bash", "-c", mgr.ExpandCmd)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("TARGETS=%s", strings.Join(targets, " ")),
			)

			expansion, err := cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("Unable to expand targets for manager %s", mgrName)
			}

			packageNames = strings.Split(string(expansion), "\n")
		} else {
			packageNames = targets
		}

		if _, ok := targetPkgs[mgr.Namespace]; !ok {
			targetPkgs[mgr.Namespace] = make(map[string]string)
		}

		for _, pkgName := range packageNames {
			if _, ok := targetPkgs[mgr.Namespace][pkgName]; ok {
				return nil, fmt.Errorf("Multiple packages with same name in %s: %s", mgr.Namespace, pkgName)
			}

			targetPkgs[mgr.Namespace][pkgName] = mgrName
		}
	}

	return targetPkgs, nil
}

// InstallPackages installs the packages given a map from sources to the packages to install from with that source
func InstallPackages(cfg *Config, pkgs map[string][]string, dryRun bool) error {
	for mgrName, mgr := range cfg.Managers {
		toInstall, ok := pkgs[mgrName]
		if !ok {
			continue
		}
		if dryRun {
			fmt.Printf("%s %s", mgr.InstallCmd, strings.Join(toInstall, " "))
		} else {
			cmd := exec.Command("bash", "-c", mgr.InstallCmd)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PKGS=%s", strings.Join(toInstall, " ")), // ignored
			)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// RemovePackages removes the packages given a config and a map from package lists to the package names to remove
func RemovePackages(cfg *Config, pkgs map[string][]string, dryRun bool) error {
	for nsName, ns := range cfg.Namespaces {
		toRemove, ok := pkgs[nsName]
		if !ok {
			continue
		}

		if dryRun {
			fmt.Printf("%s %s", ns.RemoveCmd, strings.Join(toRemove, " "))
		} else {
			cmd := exec.Command("bash", "-c", ns.RemoveCmd)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PKGS=%s", strings.Join(toRemove, " ")), // ignored
			)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

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

// GetInstalledPackages runs the appropriate shell commands to retrieve all of the packages installed in a given pkglist
func GetInstalledPackages(pkglist PkgList) (map[string]struct{}, error) {

	out, err := exec.Command("bash", "-c", pkglist.ListCmd).Output()
	if err != nil {
		return nil, fmt.Errorf("Error executing ListCmd for `%s`: %w", pkglist.Name, err)
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

	targetPkgs := make(map[string](map[string]string))

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

		defaultSrc := cfg.GetDefaultSource()

		src := ""
		for _, ident := range identRe.FindAllString(contentsNoComments, -1) {
			if ident[0] == '(' && ident[len(ident)-1] == ')' {
				if src != "" {
					return fmt.Errorf("Cannot have 2 adjacent sources: (%s) and (%s)", src, ident[1:len(ident)-1])
				}

				src = ident[1 : len(ident)-1]
			} else {
				if src == "" && defaultSrc.Name != "" {
					src = defaultSrc.Name
				}
				srcObj, ok := cfg.GetSource(src)
				if !ok {
					return fmt.Errorf("No such source found: %s", src)
				}

				if _, ok := targetPkgs[srcObj.PkgList]; !ok {
					targetPkgs[srcObj.PkgList] = make(map[string]string)
				}

				if _, ok := targetPkgs[srcObj.PkgList][ident]; ok {
					return fmt.Errorf("Multiple packages with same name in %s: %s", srcObj.PkgList, ident)
				}

				targetPkgs[srcObj.PkgList][ident] = src
				src = ""
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return targetPkgs, nil
}

// InstallPackages installs the packages given a map from sources to the packages to install from with that source
func InstallPackages(cfg *Config, pkgs map[string][]string, dryRun bool) error {
	for _, src := range cfg.Sources {
		toInstall, ok := pkgs[src.Name]
		if !ok {
			continue
		}
		if dryRun {
			fmt.Printf("%s %s", src.InstallCmd, strings.Join(toInstall, " "))
		} else {
			cmd := exec.Command("bash", "-c", src.InstallCmd)
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
	for _, pkglist := range cfg.PkgLists {
		toRemove, ok := pkgs[pkglist.Name]
		if !ok {
			continue
		}

		if dryRun {
			fmt.Printf("%s %s", pkglist.RemoveCmd, strings.Join(toRemove, " "))
		} else {
			cmd := exec.Command("bash", "-c", pkglist.RemoveCmd)
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

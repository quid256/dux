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
package cmd

import (
	"fmt"
	"path"

	"github.com/quid256/dux/utils"
	"github.com/spf13/cobra"

	"github.com/Songmu/prompter"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronizes the package state with what's in the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := utils.ConfigFromViper()
		if err != nil {
			fmt.Println(err)
			return
		}

		targets, err := utils.GetTargetPackages(cfg, path.Join(cfgDir, "pkgs"))
		if err != nil {
			fmt.Println(err)
			return
		}

		toInstall := make(map[string][]string)
		toRemove := make(map[string][]string)

		for _, pkglist := range cfg.PkgLists {
			installed, err := utils.GetInstalledPackages(pkglist)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Remove matches
			for k := range targets[pkglist.Name] {
				if _, ok := installed[k]; ok {
					delete(targets[pkglist.Name], k)
					delete(installed, k)
				}
			}

			srcToPkg := make(map[string][]string)
			for k, v := range targets[pkglist.Name] {
				srcToPkg[v] = append(srcToPkg[v], k)
				toInstall[v] = append(toInstall[v], k)
			}

			for k := range installed {
				toRemove[pkglist.Name] = append(toRemove[pkglist.Name], k)
			}

			if len(srcToPkg) > 0 || len(installed) > 0 {
				fmt.Printf("# PkgList %s\n", pkglist.Name)

				if len(installed) > 0 {
					fmt.Printf("Remove:")
					for k := range installed {
						fmt.Printf("  %s", k)
					}
					if len(installed) == 0 {
						fmt.Printf("  (none)")
					}
					fmt.Println()
				}

				if len(srcToPkg) > 0 {
					fmt.Print("Install:")
					for k, v := range srcToPkg {
						fmt.Println()
						fmt.Printf("  (%s)", k)
						for _, pkg := range v {
							fmt.Printf("  %s", pkg)
						}
					}
					fmt.Println()
				}

				fmt.Println()
			}
		}

		if len(toInstall) == 0 && len(toRemove) == 0 {
			fmt.Println("Nothing to do.")
			return

		}

		if !prompter.YN("Proceed?", false) {
			fmt.Println("Cancelling operation.")
			return
		}

		if len(toRemove) > 0 {
			fmt.Println("## Removing Packages")
			utils.RemovePackages(cfg, toRemove, dryRun)
		}

		if len(toInstall) > 0 {
			fmt.Println("## Installing Packages")
			utils.InstallPackages(cfg, toInstall, dryRun)
		}

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

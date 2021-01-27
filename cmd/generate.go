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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/quid256/dux/utils"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a Dux package index",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := utils.ConfigFromViper()
		if err != nil {
			fmt.Println(err)
			return
		}

		var b strings.Builder

		for nsName, ns := range cfg.Namespaces {
			var prefix string

			prefix = fmt.Sprintf("(%s) ", ns.DefaultManager)

			cmd := exec.Command("bash", "-c", ns.ListCmd)

			out, err := cmd.Output()
			if err != nil {
				fmt.Printf("Error executing ListCmd for `%s`: %v\n", nsName, err)
				return
			}

			for _, l := range strings.Split(string(out), "\n") {
				if len(l) == 0 {
					continue
				}
				fmt.Fprintf(&b, "%s%s\n", prefix, l)
			}
		}

		os.MkdirAll(path.Join(cfgDir, "pkgs/"), os.ModeDir|0755)
		genPath := path.Join(cfgDir, "pkgs/", "generated")

		if _, err := os.Stat(genPath); os.IsNotExist(err) {
			// TODO check the file mode here
			ioutil.WriteFile(genPath, []byte(b.String()), 0644)
		} else {
			fmt.Printf("Unable to generate file %s: file already exists.\n", genPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

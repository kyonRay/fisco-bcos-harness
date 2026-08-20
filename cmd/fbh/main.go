// fbh is the FISCO-BCOS team workflow harness CLI: the single seam
// through which harness skills touch Tencent smart sheet, WeCom, and
// GitHub. See the repo README for the command surface.
package main

import (
	"os"

	"github.com/kyonRay/fisco-bcos-harness/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}

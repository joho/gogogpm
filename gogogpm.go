package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

var originalShellScript = `
#!/usr/bin/env bash

set -eu

## Functions/
usage() {
cat << EOF
SYNOPSIS

    gpm leverages the power of the go get command and the underlying version
    control systems used by it to set your Go dependencies to desired versions,
    thus allowing easily reproducible builds in your Go projects.

    A Godeps file in the root of your Go application is expected containing
    the import paths of your packages and a specific tag or commit hash
    from its version control system, an example Godeps file looks like this:

    $ cat Godeps
    # This is a comment
    github.com/nu7hatch/gotrail         v0.0.2
    github.com/replicon/fast-archiver   v1.02   #This is another comment!
    github.com/nu7hatch/gotrail         2eb79d1f03ab24bacbc32b15b75769880629a865

    gpm has a companion tool, called [gvp](https://github.com/pote/gvp) which
    provides vendoring functionalities, it alters your GOPATH so every project
    has its own isolated dependency directory, it's usage is recommended.

USAGE
      $ gpm             # Same as 'install'.
      $ gpm install     # Parses the Godeps file, installs dependencies and sets
                        # them to the appropriate version.
      $ gpm version     # Outputs version information
      $ gpm help        # Prints this message
EOF
}

# Iterates over Godep file dependencies and sets
# the specified version on each of them.
set_dependencies() {
  deps=$(sed 's/#.*//;/^\s*$/d' < "Godeps") || echo ""

  while read package version; do
    (
      install_path="${GOPATH%%:*}/src/${package%%/...}"
      [[ -e "$install_path/.git/index.lock" ||
      -e "$install_path/.hg/store/lock" ||
      -e "$install_path/.bzr/checkout/lock" ]] && wait
      echo ">> Getting package "$package""
      go get -u -d "$package"
      echo ">> Setting $package to version $version"
      cd $install_path
      [ -d .hg ]  && hg update    -q      "$version"
      [ -d .git ] && git checkout -q      "$version"
      [ -d .bzr ] && bzr revert   -q -r   "$version"
    ) &
  done < <(echo "$deps")
  wait
  echo ">> All Done"
}

## /Functions
case "${1:-"install"}" in
  "version")
    echo ">> gpm v1.0.1"
    ;;
  "install")
    [[ -f "Godeps" ]] || (echo ">> Godeps file does not exist." && exit 1)
    (which go > /dev/null) || ( echo ">> Go is currently not installed or in your PATH" && exit 1)
    set_dependencies
    ;;
  "help")
    usage
    ;;
  *)
    usage
    exit 1
    ;;
esac
`

func main() {
	args := append([]string{"/dev/stdin"}, os.Args[1:]...)
	cmd := exec.Command("bash", args...)
	stdinWriter, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	io.WriteString(stdinWriter, originalShellScript)
	stdinWriter.Close()

	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		os.Exit(1)
	}
}

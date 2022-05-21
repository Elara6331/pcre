# pcre

[![Go Reference](https://pkg.go.dev/badge/go.arsenm.dev/pcre.svg)](https://pkg.go.dev/go.arsenm.dev/pcre)

This package provides a CGo-free port of the PCRE2 regular expression library. The [lib](lib) directory contains source code automatically translated from PCRE2's C source. This package wraps that code and provides an interface as close as possible to Go's stdlib [regexp](https://pkg.go.dev/regexp) package

---

## Supported GOOS/GOARCH:

- linux/amd64
- linux/386
- linux/arm64
- linux/arm
- linux/riscv64
- darwin/amd64
- darwin/arm64

More OS support is planned.

---

## How to transpile pcre2

In order to transpile pcre2, a Go and C compiler (preferably GCC) will be needed.

- First, install [ccgo](https://pkg.go.dev/modernc.org/ccgo/v3)

- Then, download the pcre source code. It can be found here: https://github.com/PCRE2Project/pcre2.

- Once downloaded, `cd` into the source directory

- Run `./configure`. If cross-compiling, provide the path to the cross-compiler in the `CC` variable, and set `--target` to the target architecture.

- When it completes, there should be a `Makefile` in the directory.

- Run `ccgo -compiledb pcre.json make`. Do not add `-j` arguments to the make command.

- Run the following command (replace items in triangle brackets):

```shell
CC=/usr/bin/gcc ccgo -o pcre2_<os>_<arch>.go -pkgname lib -trace-translation-units -export-externs X -export-defines D -export-fields F -export-structs S -export-typedefs T pcre.json .libs/libpcre2-8.a
```

- If cross-compiling, set the `CCGO_CC` variable to to path of the cross-compiler, and the `CCGO_AR` variable to the path of the cross-compiler's `ar` binary. Also, set `TARGET_GOARCH` to the GOARCH you're targeting and `TARGET_GOOS` to the OS you're targeting.

- Once the command completes, two go files will be created. One will start with `pcre2`, the other with `capi`. Copy both of these to the `lib` directory in this repo.
//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gateway/scripts/db/sqlserver/mysqlconv"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	srcDir := filepath.Join(root, "scripts", "db", "mysql")
	dstDir := filepath.Join(root, "scripts", "db", "sqlserver")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		panic(err)
	}
	srcNames, err := mysqlconv.ListSQLNames(srcDir)
	if err != nil {
		panic(err)
	}
	n := 0
	for name := range srcNames {
		raw, err := os.ReadFile(filepath.Join(srcDir, name))
		if err != nil {
			panic(err)
		}
		out := mysqlconv.Convert(string(raw), name, srcNames)
		if err := os.WriteFile(filepath.Join(dstDir, name), []byte(out), 0644); err != nil {
			panic(err)
		}
		n++
	}
	fmt.Printf("converted %d mysql scripts to %s\n", n, dstDir)
}

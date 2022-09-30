package main

import (
	"fmt"

	structLocker "github.com/r-usenko/struct-locker"
	"github.com/r-usenko/struct-locker/example/hiddenPackage"
)

func main() {
	cfg := hiddenPackage.Config{}

	errSet := structLocker.SetByMethod(&cfg, cfg.Uri, "http://localhost:8080")
	_ = structLocker.LockStruct(&cfg)
	fmt.Printf("URI changed to: %s\nError: %v\n\n", cfg.Uri(), errSet)

	errSet = structLocker.SetByMethod(&cfg, cfg.Uri, "https://google.com")
	fmt.Printf("URI not changed. Struct locked: %s\nError: %v\n\n", cfg.Uri(), errSet)
}

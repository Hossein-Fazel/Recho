package main

import (
	"fmt"
	"log/slog"

	"github.com/Hossein-Fazel/Recho/pkg"
)

func main() {
	pkg.InitLogger(slog.LevelInfo)
	fmt.Println("I'm Recho.")
}

package main

import (
	"log/slog"

	"github.com/Hossein-Fazel/Recho/cmd"
	"github.com/Hossein-Fazel/Recho/pkg"
)

func main() {
	pkg.InitLogger(slog.LevelInfo)
	pkg.Logger.Info("hello")
	cmd.Run()
}

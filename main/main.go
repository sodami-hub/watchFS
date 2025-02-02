package main

import (
	"github.com/sodami-hub/watchfs/client/watcher"
)

func main() {
	_ = watcher.NewWatcher("./")
}

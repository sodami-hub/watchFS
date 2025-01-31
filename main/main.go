package main

import (
	"os"

	"github.com/sodami-hub/watchfs/watcher"
)

func main() {
	argument := os.Args[1]

	myWatch := watcher.NewWatcher(argument)

	myWatch.Watch()
}

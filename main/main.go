package main

import (
	"os"
)

func main() {
	argument := os.Args[1]

	myWatch := watchfs.newWatcher(argument)

	myWatch.initSearch()
	myWatch.initList()
}

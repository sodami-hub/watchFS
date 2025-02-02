// 클라이언트 테스트를 위한 main 함수

package main

import (
	"github.com/sodami-hub/watchfs/client/watcher"
)

func main() {
	_ = watcher.NewWatcher("./")
}

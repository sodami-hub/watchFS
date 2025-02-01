package watcher

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

type fs struct {
	initDir     string
	directories []string
	allFiles    map[string]string
	changes     map[string]string // change 값이 변한 파일들의 슬라이스
}

func NewWatcher(dir string) fs {
	initDir := make([]string, 0, 10)
	initDir = append(initDir, string(dir))
	myWatcher := &fs{
		initDir:     dir,
		directories: initDir,
		allFiles:    make(map[string]string),
		changes:     make(map[string]string),
	}

	return *myWatcher
}

func (fs *fs) dirSearch() error {
	err := filepath.Walk(fs.initDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != fs.initDir {
				path = fs.initDir + path
				fs.directories = append(fs.directories, path)
			}
		} else {
			path = fs.initDir + path
			modTime := info.ModTime().String()
			fs.allFiles[path] = modTime
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %v", err)
	}
	return nil
}

func (fs fs) List() {
	fmt.Println("디렉토리 리스트")
	for _, dir := range fs.directories {
		fmt.Println(dir)
	}
	fmt.Println("디렉토리 / 파일:수정시간 리스트")
	for k, v := range fs.allFiles {
		fmt.Printf("filename : %s  //  modtime: %s\n", k, v)
	}
	fmt.Println("변경 파일 목록")
	for k, v := range fs.changes {
		fmt.Printf("filename : %s // 변경사항 : %s\n", k, v)
	}
}

func (fs *fs) Watch() error {

	err := fs.dirSearch()
	if err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("make fsnotify error : %v", err)
	}
	defer watcher.Close()

	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
	var isDir bool = false

	go func() {
		for {
			select {
			case <-sigChan:
				fmt.Println("인터럽트 발생 종료...")
				done <- struct{}{}
				return
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Create:
					info, err := os.Stat(event.Name)
					if err != nil {
						fmt.Println("생성된 파일 정보 가져오는 중 에러 발생")
						done <- struct{}{}
					}
					if info.IsDir() {
						fmt.Println("디렉터리 생성")
						fs.directories = append(fs.directories, event.Name)
						watcher.Add(event.Name)
					} else {
						fmt.Println("파일생성")
						fs.changes[event.Name] = "create"
					}
					fs.List()
				case fsnotify.Remove:
					for _, dir := range fs.directories {
						if dir == event.Name {
							fmt.Println("디렉터리 삭제", event.Name)
							isDir = true
							break
						}
					}
					if isDir {
						for i, dir := range fs.directories {
							if dir == event.Name {
								fs.directories = append(fs.directories[:i], fs.directories[i+1:]...)
								break
							}
						}
					} else {
						var deleteMarking bool = false
						fmt.Println("파일삭제", event.Name)
						for k := range fs.changes {
							if k == event.Name {
								delete(fs.changes, k)
								deleteMarking = true
								break
							}
						}
						if !deleteMarking {
							fs.changes[event.Name] = "delete"
						}
					}
					fs.List()
					isDir = false
				case fsnotify.Write:
					fmt.Println("파일 수정")
					for k := range fs.allFiles {
						if k == event.Name {
							fileMod, err := os.Stat(event.Name)
							if err != nil {
								fmt.Println("수정된 파일 정보 가져오기 에러")
								done <- struct{}{}
							}
							fs.changes[event.Name] = fileMod.ModTime().String()
							break
						}
					}
					fs.List()
				}
			}
		}
	}()
	for _, dir := range fs.directories {
		err := watcher.Add(dir)
		if err != nil {
			return fmt.Errorf("watcher add path error : %v", err)
		}
	}
	<-done
	return nil
}

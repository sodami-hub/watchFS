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
	allFiles    []map[string]string
	changes     []map[string]string // change 값이 변한 파일들의 슬라이스
}

func NewWatcher(dir string) fs {
	initDir := make([]string, 0, 10)
	initDir = append(initDir, string(dir))
	myWatcher := &fs{
		initDir:     dir,
		directories: initDir,
		allFiles:    make([]map[string]string, 0, 30),
		changes:     make([]map[string]string, 0, 20),
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
				fs.directories = append(fs.directories, path)
			}
		} else {
			fileMap := make(map[string]string)
			modTime := info.ModTime().String()
			fileMap[path] = modTime
			fs.allFiles = append(fs.allFiles, fileMap)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %v", err)
	}
	return nil
}

func (fs fs) InitList() {
	fmt.Println("디렉토리 리스트")
	for _, dir := range fs.directories {
		fmt.Println(dir)
	}
	fmt.Println("디렉토리 / 파일:수정시간 리스트")
	for _, dir := range fs.allFiles {
		for k, v := range dir {
			fmt.Printf("filename : %s  //  modtime: %s\n", k, v)
		}
	}
	fmt.Println("변경 파일 목록")
	for _, dir := range fs.changes {
		for k, v := range dir {
			fmt.Printf("filename : %s // 변경사항 : %s\n", k, v)
		}
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
					fmt.Println("파일생성", event.Name)
					info, err := os.Stat(event.Name)
					if err != nil {
						fmt.Println("생성된 파일 정보 가져오는 중 에러 발생")
						done <- struct{}{}
					}
					if info.IsDir() {
						fs.directories = append(fs.directories, info.Name())
						watcher.Add(info.Name())
					} else {
						//modTime := info.ModTime()
						fileMap := map[string]string{event.Name: "create"}
						fs.changes = append(fs.changes, fileMap)
					}
					fs.InitList()
				case fsnotify.Remove:
					for _, dir := range fs.directories {
						if dir == event.Name {
							fmt.Println("디렉터리 삭제", event.Name)
							// 디렉터리를 삭제할 경우 해당 디렉터리가 포함된 모든 파일, 디렉터리에 대해서 delete 로 바꾸기
						} else {
							fmt.Println("파일삭제", event.Name)
							fileMap := make(map[string]string)
							fileMap[event.Name] = "delete"
							fs.changes = append(fs.allFiles, fileMap)
						}
					}
					fs.InitList()
				case fsnotify.Write:
					fmt.Println("수정")
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

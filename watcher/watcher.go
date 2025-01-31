package watcher

import (
	"fmt"
	"os"
	"path/filepath"
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

func (fs *fs) InitSearch() error {
	err := filepath.Walk(fs.initDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			fs.directories = append(fs.directories, path)
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
}

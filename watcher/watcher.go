package watcher

import (
	"fmt"
	"os"
)

type dirAndFiles struct {
	dir   string
	files []map[string]string
}

type fs struct {
	directories []string
	allFiles    []dirAndFiles
	changes     []map[string]string // change 값이 변한 파일들의 슬라이스
}

func newWatcher(dir string) fs {
	initDir := make([]string, 10)
	initDir = append(initDir, string(dir))
	daf := make([]dirAndFiles, 10)
	myWatcher := &fs{
		directories: initDir,
		allFiles:    daf,
		changes:     make([]map[string]string, 20),
	}

	return *myWatcher
}

func (fs *fs) initSearch() error {

	for _, directory := range fs.directories {
		innerDir, err := os.ReadDir(directory)
		if err != nil {
			return fmt.Errorf("ReadDir error : %v", err)
		}
		daf := &dirAndFiles{
			dir: directory,
		}
		fileMap := make(map[string]string)
		for _, innerFile := range innerDir {
			if innerFile.IsDir() {
				fs.directories = append(fs.directories, innerFile.Name())
			} else {
				info, err := innerFile.Info()
				if err != nil {
					return fmt.Errorf("FileInfo error : %v", err)
				}
				modTime := info.ModTime()
				fileMap[innerFile.Name()] = modTime.String()
				daf.files = append(daf.files, fileMap)
			}
		}
		fs.allFiles = append(fs.allFiles, *daf)
	}
	return nil
}

func (fs fs) initList() {
	fmt.Println("디렉토리 리스트")
	for _, dir := range fs.directories {
		fmt.Println(dir)
	}
	fmt.Println("디렉토리 / 파일:수정시간 리스트")
	for _, dir := range fs.allFiles {
		fmt.Println("====", dir.dir, "====")
		for _, file := range dir.files {
			for k, v := range file {
				fmt.Printf("filename : %s : modtime: %s\n", k, v)
			}
		}
	}
}

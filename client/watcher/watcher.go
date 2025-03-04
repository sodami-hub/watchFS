package watcher

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	api "github.com/sodami-hub/watchfs/api/v1"
	"google.golang.org/protobuf/proto"
)

type FS struct {
	initDir     string
	directories []string
	allFiles    map[string]string
	changes     map[string]string // change 값이 변한 파일들의 슬라이스
}

func NewWatcher(dir string) (FS, error) {
	dirs := make([]string, 0, 10)
	dirs = append(dirs, string(dir))
	myWatcher := &FS{
		initDir:     dir,
		directories: dirs,
		allFiles:    make(map[string]string),
		changes:     make(map[string]string),
	}

	return *myWatcher, nil
}

// 클라이언트의 root디렉터리 하위의 파일 시스템을 검색해서 fs 구조체를 전달하는 함수
func searchDir(fs *FS) error {
	if fs.initDir == "" {
		fs.initDir = "./"
	}
	err := filepath.Walk(fs.initDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		bool := strings.Contains(path, ".garage")
		if info.IsDir() {
			if path != fs.initDir && !bool {
				path = fs.initDir + path
				fs.directories = append(fs.directories, path)
			}
		} else {
			if !bool {
				path = fs.initDir + path
				modTime := info.ModTime().String()
				fs.allFiles[path] = modTime
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %v", err)
	}
	return nil
}

func (fs *FS) loadFS() error {
	file, err := os.Open("./.garage/clientFS")
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		err = searchDir(fs)
		if err != nil {
			return fmt.Errorf("error walking the path %v", err)
		}
		err = os.MkdirAll("./.garage", 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}

		file, err := os.OpenFile("./.garage/clientFS", os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("file create error : %v", err)
		}
		_ = file.Close()
	} else {
		fmt.Println("설정파일 있다.")
		// 설정 파일이 있는 경우 저장된 파일 시스템과 현재 클라이언트의 파일 시스템의 상태가 같은지를 확인해야 된다.

		// 저장된 파일시스템 정보를 가져온다.
		var savedfs api.ClientFS
		b, err := os.ReadFile("./.garage/clientFS")
		if err != nil {
			return err
		}
		err = proto.Unmarshal(b, &savedfs)
		if err != nil {
			return err
		}

		// 현재 클라이언트의 파일시스템 정보를 가져온다.
		searchDir(fs)

		savedDir := savedfs.Directories
		savedAllFiles := savedfs.AllFiles
		savedChanges := savedfs.Changes
		if savedChanges == nil {
			savedChanges = make(map[string]string)
		}
		// 저장된 데이터와 현재 클라이언트 파일시스템 정보를 비교해서 데이터를 수정한다.
		for _, now := range fs.directories {
			if !sliceContains(savedDir, now) {
				savedDir = append(savedDir, now)
			}
		}

		for i, v := range savedDir {
			if !sliceContains(fs.directories, v) {
				savedDir = append(savedDir[:i], savedDir[i+1:]...)
			}
		}

		for k, v := range fs.allFiles {
			if fsVal, exists := savedAllFiles[k]; exists {
				if v != fsVal {
					savedChanges[k] = v
				}
			} else {
				savedChanges[k] = "create"
			}
		}

		for k := range savedAllFiles {
			if _, exists := fs.allFiles[k]; !exists {
				savedChanges[k] = "delete"
			}
		}

		for k, v := range savedChanges {
			if v == "create" {
				if _, exists := fs.allFiles[k]; !exists {
					delete(savedChanges, k)
				}
			}
		}

		fs.initDir = savedfs.InitDir
		fs.allFiles = savedAllFiles
		fs.directories = savedDir
		fs.changes = savedChanges
	}
	return nil
}

func sliceContains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (fs *FS) saveFS() error {
	file, err := os.OpenFile("./.garage/clientFS", os.O_WRONLY|os.O_CREATE, 0644)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return err
	}
	b, err := proto.Marshal(&api.ClientFS{
		InitDir:     fs.initDir,
		Directories: fs.directories,
		AllFiles:    fs.allFiles,
		Changes:     fs.changes,
	})
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	return err
}

func (fs FS) Status() {
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
		fmt.Printf("changes : %s // 변경사항 : %s\n", k, v)
	}
}

func (fs FS) ChangeFile() {
	fmt.Println("변경 사항")
	for k, v := range fs.changes {
		fmt.Printf("[%s : %s]\n", k, v)
	}
}

func (fs *FS) Watch() error {

	err := fs.loadFS()
	if err != nil {
		return err
	}
	fs.saveFS()
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
				// 인터럽트 발생 종료
				done <- struct{}{}
				return
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
				fs.saveFS()
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Create:
					info, err := os.Stat(event.Name)
					if err != nil {
						// 생성된 파일 정보 가져오는 중 에러 발생
						fs.saveFS()
						done <- struct{}{}
					}
					if info.IsDir() {
						// 디렉터리 생성
						path := event.Name
						fs.directories = append(fs.directories, path)
						watcher.Add(event.Name)
					} else {
						// 파일생성
						if bool := strings.HasPrefix(event.Name, "./"); !bool {
							event.Name = "./" + event.Name
						}
						fs.changes[event.Name] = "create"
					}
					fs.saveFS()
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
								fs.saveFS()
								break
							}
						}
					} else {
						// 파일삭제
						var deleteMarking bool = false
						for k := range fs.changes {
							if k == event.Name {
								delete(fs.changes, k)
								fs.saveFS()
								deleteMarking = true
								break
							}
						}
						if !deleteMarking {
							fs.changes[event.Name] = "delete"
							fs.saveFS()
						}
					}
					fs.saveFS()
					isDir = false
				case fsnotify.Write:
					if bool := strings.HasPrefix(event.Name, "./"); !bool {
						event.Name = "./" + event.Name
					}
					for k := range fs.allFiles {
						if k == event.Name {
							fileMod, err := os.Stat(event.Name)
							if err != nil {
								fmt.Println("수정된 파일 정보 가져오기 에러")
								done <- struct{}{}
							}
							fs.changes[event.Name] = fileMod.ModTime().String()
							fs.saveFS()
							break
						}
					}
					fs.saveFS()
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

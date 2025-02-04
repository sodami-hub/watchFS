// 클라이언트 테스트를 위한 main 함수

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	api "github.com/sodami-hub/watchfs/client/api/v1"
	"github.com/sodami-hub/watchfs/client/garage"
	"google.golang.org/protobuf/proto"
)

var childProcess *os.Process

func main() {
	args := os.Args[1:]
	userInfo := &api.UserInfo{}
	var hasUserInfo bool
	f, err := os.OpenFile(".garage/.user", os.O_RDWR, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		if err == os.ErrNotExist {
			hasUserInfo = false
		}
	} else {
		hasUserInfo = true
		b := make([]byte, 1024)
		n, err := f.Read(b)
		if err != nil {
			fmt.Println(err)
			return
		}
		proto.Unmarshal(b[:n], userInfo)
	}

	switch args[0] {
	case "conn":
		if hasUserInfo {
			// 설정파일이 있고 garage start 명령을 입력하면 자식 쉘에서 감시를 시작한다.
			cmd := exec.Command("go", "run", "main.go", "start")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // 새로운 프로세스 그룹 생성
			err := cmd.Start()
			if err != nil {
				fmt.Println(err)
				return
			}
			childProcess = cmd.Process
			fmt.Printf("Started child process with PID %d\n", childProcess.Pid)
			userInfo.ChildProcessPid = int32(childProcess.Pid)
			b, err := proto.Marshal(userInfo)
			if err != nil {
				fmt.Println(err)
				return
			}
			_, err = f.Write(b)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = garage.GarageConn(args[1], args[2])
			if err != nil {
				fmt.Println(err)
				return
			}

		}
	case "init":
		_ = garage.GarageInit(args[1])
	case "start":
		_ = garage.GarageWatch(userInfo)
	case "stop":
		if userInfo.ChildProcessPid != 0 {
			pgid := -int(userInfo.ChildProcessPid)
			err := syscall.Kill(pgid, syscall.SIGTERM)
			if err != nil {
				fmt.Println("Failed to stop child process:", err)
				return
			} else {
				fmt.Println("Child process stopped")
			}
		} else {
			fmt.Println("No child process to stop")
		}
	case "changes":
		_ = garage.ChangeFile()
	case "all":
		_ = garage.All()
	}
}

// 클라이언트 테스트를 위한 main 함수

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	api "github.com/sodami-hub/watchfs/api/v1"
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
	case "join":
		if len(args) != 3 {
			fmt.Println("garage join [id] [pw] // garage 서비스에 가입하세요.")
			return
		}
		err := garage.GarageJoin(args[1], args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("회원가입 완료")
	case "conn":
		if hasUserInfo {
			if len(args) > 1 {
				fmt.Println("이미 등록된 계정정보 있습니다. garage conn 으로 저장된 garage에 접속하거나, garage를 생성하세요.")
				return
			}
			file, err := os.OpenFile(".garage/.user", os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = StartWatch(file, userInfo)
			if err != nil {
				fmt.Println(err)
				return
			}
			_ = file.Close()
		} else {
			if len(args) != 3 {
				fmt.Println("garage conn [사용자 id] [사용자 password] // 현재 디렉터리에 사용자 정보를 설정하세요.")
				return
			}
			err = garage.GarageConn(args[1], args[2])
			if err != nil {
				fmt.Println(err)
				return
			}

		}
	case "init":
		if len(args) > 2 {
			fmt.Println("init [garage_name] garage name은 띄어쓰기를 허용하지 않습니다.")
			return
		}
		err = garage.GarageInit(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
	case "start":
		err = garage.GarageWatch(userInfo)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "stop":
		if len(args) > 1 {
			fmt.Println("stop 옵션에는 추가 실행인자가 필요하지 않습니다.")
			return
		}
		err := StopProc(int(userInfo.ChildProcessPid))
		if err != nil {
			fmt.Println(err)
			return
		}
	case "changes":
		if len(args) > 1 {
			fmt.Println("change 옵션에는 추가 실행인자가 필요하지 않습니다.")
			return
		}
		err = garage.ChangeFile()
		if err != nil {
			fmt.Println(err)
			return
		}
	case "all":
		if len(args) > 1 {
			fmt.Println("all 옵션에는 추가 실행인자가 필요하지 않습니다.")
			return
		}
		err = garage.All()
		if err != nil {
			fmt.Println(err)
			return
		}
	case "save": // 로컬의 변경사항을 리모트에 저장하기 위해서 변경 내용을 저장(commit)
		if len(args) > 2 {
			fmt.Println("save [msg] // save옵션과 message를 사용한다.")
			return
		}
		err = garage.Save(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}

		err = StopProc(int(userInfo.ChildProcessPid))
		if err != nil {
			fmt.Println(err)
			return
		}

		err = os.Remove(".garage/clientFS")
		if err != nil {
			fmt.Println(err)
			return
		}

		file, err := os.OpenFile(".garage/.user", os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = StartWatch(file, userInfo)
		if err != nil {
			fmt.Println(err)
			return
		}
		_ = file.Close()
	case "history":
		if len(args) > 1 {
			fmt.Println("history 옵션에는 추가 실행인자가 필요하지 않습니다.")
			return
		}
		err = garage.ShowHistory()
		if err != nil {
			fmt.Println(err)
			return
		}
	case "push":
		if len(args) > 1 {
			fmt.Println("push 옵션에는 추가 실행인자가 필요하지 않습니다.")
			return
		}
		err := garage.Push()
		if err != nil {
			fmt.Println(err)
			return
		}

	}
}

func StopProc(pgid int) error {
	if pgid != 0 {
		pgid := -pgid // 생성된 프로세스를 음수로 바꿔서 그룹 전체에 시그널을 보냄
		err := syscall.Kill(pgid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("Failed to stop child process: ", err)
			return err
		} else {
			fmt.Println("Child process stopped")
		}
	} else {
		fmt.Println("No child process to stop")
	}
	return nil
}

func StartWatch(file *os.File, userInfo *api.UserInfo) error {

	// 설정파일이 있고 garage start 명령을 입력하면 자식 쉘에서 감시를 시작한다.
	cmd := exec.Command("go", "run", "client.go", "start")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // 새로운 프로세스 그룹 생성
	err := cmd.Start()
	if err != nil {
		return err
	}
	childProcess = cmd.Process
	fmt.Printf("Started child process with PID %d\n", childProcess.Pid)
	userInfo.ChildProcessPid = int32(childProcess.Pid)
	b, err := proto.Marshal(userInfo)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(2 * time.Second)
	_, err = os.Stat(".garage/history/historySeq")
	if err != nil {
		fmt.Println("초기 데이터를 서버에 저장합니다.")
		garage.FirstUpload()
	}
	return nil
}

// 클라이언트의 정보를 저장 클라이언트 측 api 제공

package garage

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	api "github.com/sodami-hub/watchfs/client/api/v1"
	"github.com/sodami-hub/watchfs/client/watcher"
	"google.golang.org/protobuf/proto"
)

// protobuf 사용으로 구조체를 사용하지 않개됨....
type UserInfo struct {
	id              string
	pw              string
	garageName      string
	childProcessPid int32
}

// 회원 가입
/*
생성돌 gRPC 서버에 post 요청으로 id/pw를 보내서 회원을 가입한다.
그리고 garage 클라이언트 실행 파일을 /usr/bin/ 으로 복사한다.
*/

/*
서비스 접속 -
해당 디렉터리에 설정 정보가 있으면 garage conn 명령으로 해당 설정 정보를 불러와서
서버에 접속하고 디렉터리 감시를 시작함 - GarageWatch()
설정 정보가 없으면 garage conn id pw 명령으로 서버에 접속하고, - GarageConn()
garage init garageName 으로 감시 디렉터리 추가하고 감시 시작, - GarageInit() -> GarageWatch()
*/

// $ garage conn id pw
func GarageConn(id, pw string) error {

	userInfo := &api.UserInfo{
		Id: id,
		Pw: pw,
	}

	b, err := proto.Marshal(userInfo)
	if err != nil {
		return err
	}
	// 서버 접속 - 서버에서 id - pw 확인!!!!
	err = os.MkdirAll("./.garage", 0755)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(".garage/.user", os.O_CREATE|os.O_WRONLY, 0644)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// $ garage init garageName
func GarageInit(garageName string) error {
	// 유저 정보 가져오기
	user := &api.UserInfo{}
	err := LoadUserInfo(user)
	if err != nil {
		return err
	}
	user.GarageName = garageName

	file, err := os.OpenFile(".garage/.user", os.O_RDWR|os.O_TRUNC, 0644)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return err
	}
	// ToDo : 서버로 데이터 보내서 사용자 확인하고 서버에 userId/[garageName] 디렉터리 생성

	protoM, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	_, err = file.Write(protoM)
	if err != nil {
		return err
	}
	return nil
}

// & garage start
func GarageWatch(user *api.UserInfo) error {

	// ToDo : user를 서버에 보내서 사용자 및 레포지토리 확인!

	myWatcher, err := watcher.NewWatcher("./")
	if err != nil {
		return err
	}

	myWatcher.Watch()

	return nil
}

func ChangeFile() error {
	myFS := &api.ClientFS{}
	err := LoadClientFS(myFS)
	if err != nil {
		return err
	}
	changeFile := myFS.Changes
	fmt.Println("[변경 된 파일들]")
	for k, v := range changeFile {
		fmt.Printf("[%s : %s]\n", k, v)
	}
	return nil
}

func All() error {
	myFS := &api.ClientFS{}
	err := LoadClientFS(myFS)
	if err != nil {
		return err
	}
	dirs := myFS.Directories
	fmt.Println("[디렉토리 목록]")
	for _, v := range dirs {
		fmt.Printf("[%s]\n", v)
	}

	allFile := myFS.AllFiles
	fmt.Println("\n[서버 저장 대기중인 파일들]")
	for k, v := range allFile {
		fmt.Printf("[%s : %s]\n", k, v)
	}
	changeFile := myFS.Changes
	fmt.Println("\n[변경 된 파일들]")
	for k, v := range changeFile {
		fmt.Printf("[%s : %s]\n", k, v)
	}
	return nil
}

type saveChanges struct {
	seq         uint32
	msg         string
	changeOrder map[string]string
}

type changeHistory struct {
	history []saveChanges
}

// github의 커밋과 같은 역할 로컬의 변경사항을 메시지와 함께 리모트에 푸시할 상태로 업데이트한다.
// 아직 리모트에 푸쉬되는 것은 아님. 상태만 저장
func Save(msg string) error {
	// 클라이언트의 파일시스템 정보 가져오기
	myFS := &api.ClientFS{}
	err := LoadClientFS(myFS)
	if err != nil {
		return err
	}
	// 기존의 history가 있는지 확인
	f, err := os.Open(".garage/history/history")
	defer func() {
		_ = f.Close()
	}()

	if err != nil {
		// 없으면 첫번째 히스토리 생성해서 파일에 저장
		err := os.MkdirAll(".garage/history", 0755)
		if err != nil {
			return err
		}
		saveInfo := &api.SaveChanges{
			Msg:         msg,
			Seq:         0,
			ChangeOrder: myFS.Changes,
		}
		histories := make([]*api.SaveChanges, 0)
		histories = append(histories, saveInfo)
		history := &api.ChangeHistory{
			History: histories,
		}

		hisFile, err := os.OpenFile(".garage/history/myHistory", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		b, err := proto.Marshal(history)
		if err != nil {
			return err
		}
		_, err = hisFile.Write(b)
		if err != nil {
			return err
		}
		// 수정/생성 된 파일을 다른 디렉터리로 이동(나중에 롤백할 수 있도록)
		err = MoveChangedFile(saveInfo.Seq)
		if err != nil {
			return err
		}

	} else {
		// 기존의 history가 있는경우

	}
	return nil
}

func MoveChangedFile(seq uint32) error {
	myFS := &api.ClientFS{}
	err := LoadClientFS(myFS)

	if err != nil {
		return err
	}
	savePath := ".garage/history/changeOrder_" + strconv.Itoa(int(seq))
	err = os.Mkdir(savePath, 0755)
	if err != nil {
		return err
	}
	// 새롭게 생성된 디렉터리 만들기
	for i, dir := range myFS.Directories {

		if i == 0 {
			continue
		}
		dir = dir[1:]
		err := os.MkdirAll(savePath+dir, 0755)
		if err != nil {
			return err
		}
	}

	for k := range myFS.Changes {
		file, err := os.OpenFile(k, os.O_RDONLY, 0655)
		if err != nil {
			return err
		}
		if bool := strings.HasPrefix(k, "."); bool {
			k = k[1:]
		} else {
			k = "/" + k
		}
		savedFile, err := os.OpenFile(savePath+k, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		buf := make([]byte, 4096)
		for {
			n1, err := file.Read(buf)
			if n1 == 0 {
				break
			}
			if err != nil {
				return err
			}
			_, err = savedFile.WriteString(string(buf[:n1]))
			if err != nil {
				return err
			}
		}
		_ = file.Close()
		_ = savedFile.Close()
	}
	return nil
}

// 클라이언트의 정보를 저장 클라이언트 측 api 제공

package garage

import (
	"fmt"
	"os"

	api "github.com/sodami-hub/watchfs/client/api/v1"
	"github.com/sodami-hub/watchfs/client/watcher"
	"google.golang.org/protobuf/proto"
)

// protobuf 사용으로 구조체를 사용하지 않개됨....
// type UserInfo struct {
// 	id         string
// 	pw         string
// 	garageName string
// 	supervisor watcher.FS
// }

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
	f, err := os.Open(".garage/.user")
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("현재 디렉터리에서 사용자 인증이 필요하다. \n garage conn id pw \n %v", err)
	}
	b := make([]byte, 1024)
	n, err := f.Read(b)
	if err != nil {
		return err
	}
	user := &api.UserInfo{}
	proto.Unmarshal(b[:n], user)
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
	f, err := os.Open(".garage/clientFS")
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	buf := make([]byte, 2048)
	n, err := f.Read(buf)
	if err != nil {
		return err
	}
	proto.Unmarshal(buf[:n], myFS)
	changeFile := myFS.Changes
	fmt.Println("[변경 된 파일들]")
	for k, v := range changeFile {
		fmt.Printf("[%s : %s]\n", k, v)
	}
	return nil
}

func All() error {
	myFS := &api.ClientFS{}
	f, err := os.Open(".garage/clientFS")
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	buf := make([]byte, 2048)
	n, err := f.Read(buf)
	if err != nil {
		return err
	}
	proto.Unmarshal(buf[:n], myFS)
	allFile := myFS.AllFiles
	fmt.Println("[서버 저장 대기중인 파일들]")
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

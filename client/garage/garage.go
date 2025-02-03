// 클라이언트의 정보를 저장 클라이언트 측 api 제공

package garage

import (
	"os"

	"github.com/sodami-hub/watchfs/client/watcher"
)

type userInfo struct {
	id         string
	pw         string
	garageName string
}

// 회원 가입
func SignIn(id, pw string) error {

	// garage 서버 접속

	user := &userInfo{
		id: id,
		pw: pw,
	}

	// protobuf UserInfo 필드에 user데이터 저장
	// UserInfo 필드를 직렬화해서(marshal) 서버로 보내고 아이디 중복 확인 등 문제 없으면
	// 회원가입 성공

	return nil
}

/*
서비스 접속 -
해당 디렉터리에 설정 정보가 있으면 garage conn 명령으로 해당 설정 정보를 불러와서
서버에 접속하고 디렉터리 감시를 시작함 - GarageWatch()
설정 정보가 없으면 garage conn id pw 명령으로 서버에 접속하고, - GarageConn()
garage init garageName 으로 감시 디렉터리 추가하고 감시 시작, - GarageInit() -> GarageWatch()
*/
func GarageConn(id, pw string) error {
	user := &userInfo{
		id: id,
		pw: pw,
	}

	// protobuf UserInfo 필드에 user데이터 저장
	// 서버 접속 - 서버에서 id - pw 확인 후 접속

	// user정보 파일 생성
	f, err := os.OpenFile(".garage/.user", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	// UserInfo 필드를 직렬화해서(marshal) f 에 저장

	return nil
}

func GarageInit(garageName string) error {
	// 유저 정보 가져오기
	f, err := os.OpenFile(".garage/.user", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	// f 언마샬로 protobuf 타입으로 언마샬
	// garageName 필드에 데이터 추가

	// 서버로 데이터 보내서 서버에 userId/[garageName] 디렉터리 생성

	// 감시 시작!
	return watcher.NewWatcher("./")
}

func GarageWatch() error {
	return watcher.NewWatcher("./")
}

package garage

import (
	"fmt"
	"os"

	api "github.com/sodami-hub/watchfs/api/v1"
	"google.golang.org/protobuf/proto"
)

func LoadClientFS(fs *api.ClientFS) error {
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
	err = proto.Unmarshal(buf[:n], fs)
	return err
}

func LoadUserInfo(user *api.UserInfo) error {
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
	err = proto.Unmarshal(b[:n], user)
	return err
}

func LoadHistorySeq(seq *api.HistorySeq) error {
	seqFile, err := os.OpenFile(".garage/history/historySeq", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = seqFile.Close()
	}()

	buf := make([]byte, 1024)
	n, err := seqFile.Read(buf)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(buf[:n], seq)
	if err != nil {
		return err
	}
	return nil
}

// push를 하기 위해서 HistorySeq 구조체의 uploadSeq를 불러와서 upload된 위치를 확인하고 그 이후부터의 값을 전달한다.
// garage.push() 함수에서 사용하기 위한 함수이다.
func LoadSaveChanges(saveInfo *api.SaveChanges, seq int) error {

	path := fmt.Sprintf(".garage/history/changeOrder_%d/save_%d", seq, seq)
	info, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(info, saveInfo)
	return err
}

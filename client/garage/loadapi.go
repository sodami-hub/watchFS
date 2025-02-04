package garage

import (
	"fmt"
	"os"

	api "github.com/sodami-hub/watchfs/client/api/v1"
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
	proto.Unmarshal(buf[:n], fs)
	return nil
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
	proto.Unmarshal(b[:n], user)
	return nil
}

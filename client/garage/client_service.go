// 클라이언트의 정보를 저장 클라이언트 측 api 제공

package garage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	api "github.com/sodami-hub/watchfs/api/v1"
	"github.com/sodami-hub/watchfs/client/watcher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

// protobuf 사용으로 구조체를 사용하지 않개됨....
// type UserInfo struct {
// 	id              string
// 	pw              string
// 	garageName      string
// 	childProcessPid int32
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

var addr, caCertFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "server address")
	flag.StringVar(&caCertFn, "ca-cert", "cert.pem", "CA certificate")
}

func serverConn() (*grpc.ClientConn, api.GarageClient, context.Context, error) {
	flag.Parse()

	caCert, err := os.ReadFile(caCertFn)
	if err != nil {
		return nil, nil, nil, err
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, nil, nil, fmt.Errorf("failed to add certificate to pool")
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					CurvePreferences: []tls.CurveID{tls.CurveP256},
					MinVersion:       tls.VersionTLS12,
					RootCAs:          certPool,
					NextProtos:       []string{"h2"}, // ALPN(Application-Layer Protocol Negotiation) 속성 설정
				},
			),
		),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	garageClient := api.NewGarageClient(conn)
	ctx := context.Background()
	return conn, garageClient, ctx, nil
}

// $ garage join id pw
// garage servie 회원가입
func GarageJoin(id, pw string) error {
	// 회원 가입 후 서버 접속 종료
	userInfo := &api.UserInfo{
		Id: id,
		Pw: pw,
	}

	conn, garageClient, ctx, err := serverConn()
	if err != nil {
		return err
	}

	response, err := garageClient.Join(ctx, userInfo)
	if err != nil {
		return err
	}
	fmt.Println(response)
	conn.Close()

	return nil
}

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

	conn, garageClient, ctx, err := serverConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	response, err := garageClient.LogIn(ctx, userInfo)
	if err != nil {
		return err
	}
	fmt.Println(response)

	err = os.MkdirAll("./.garage", 0755)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(".garage/.user", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
/*
추가 로직
-> 같은 garage 이름으로 초기화하면 기존의 서버에 저장된 내용을 복사한다.
-> 다른 위치에서 같은 garage name을 사용할 때 동기화가 가능하도록 업데이트 필요
*/
func GarageInit(garageName string) error {
	// 유저 정보 가져오기
	user := &api.UserInfo{}
	err := LoadUserInfo(user)
	if err != nil {
		return err
	}
	if user.GarageName != "" {
		return fmt.Errorf("이미 설정된 garage가 존재합니다. 다른 디렉터리에 새롭게 설정하세요")
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
	conn, garageClient, ctx, err := serverConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	response, err := garageClient.InitGarage(ctx, user)
	if err != nil {
		fmt.Println(err)
		user.GarageName = ""
		userInfo, err := proto.Marshal(user)
		if err != nil {
			return err
		}

		_, err = file.Write(userInfo)
		if err != nil {
			return err
		}
		return err
	}
	fmt.Println(response)

	userInfo, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	_, err = file.Write(userInfo)
	if err != nil {
		return err
	}
	return nil
}

func FirstUpload() error {
	conn, garageClient, ctx, err := serverConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	fs := &api.ClientFS{}

	err = LoadClientFS(fs)
	if err != nil {
		return err
	}

	user := &api.UserInfo{}
	err = LoadUserInfo(user)
	if err != nil {
		return err
	}

	rootDir := fmt.Sprintf("./root/%s_%s", user.Id, user.GarageName)
	fmt.Println("초기 디렉터리 정보를 서버로 푸쉬합니다.", rootDir)
	var endFlag bool = false
	count := 1
	for k, v := range fs.AllFiles {
		if count == len(fs.AllFiles) {
			endFlag = true
		}

		fd, err := os.ReadFile(k)
		if err != nil {
			return err
		}

		file := &api.File{
			RootDir:  rootDir,
			FilePath: k, //rootDir + "/" + i[2:]
			Desc:     v,
			FileData: fd,
			EndFile:  endFlag,
		}
		response, err := garageClient.UploadFiles(ctx, file)
		if err != nil {
			return err
		}
		count++
		fmt.Println(response)
	}

	newSeq := &api.HistorySeq{
		Seq:       0,
		UploadSeq: 0,
	}
	err = os.MkdirAll(".garage/history", 0755)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(".garage/history/historySeq", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()
	b, err := proto.Marshal(newSeq)
	if err != nil {
		return err
	}

	_, err = fd.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func CheckUser(user *api.UserInfo) error {
	conn, garageClient, ctx, err := serverConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	response, err := garageClient.LogIn(ctx, user)
	if err != nil {
		return err
	}
	fmt.Println(response)
	return nil
}

// & garage start
func GarageWatch(user *api.UserInfo) error {

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
	fmt.Println("\n[저장된 파일들 // 서버 저장 상태를 확인하려면 history 옵션 사용]")
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

// type saveChanges struct {
// 	seq         uint32
// 	msg         string
// 	changeOrder map[string]string
// }

// type changeHistory struct {
// 	seq     uint32
// 	history []saveChanges
// }

// github의 커밋과 같은 역할 로컬의 변경사항을 메시지와 함께 리모트에 푸시할 상태로 업데이트한다.
// 아직 리모트에 푸쉬되는 것은 아님. 상태만 저장
func Save(msg string) error {
	// 클라이언트의 파일시스템 정보 가져오기
	myFS := &api.ClientFS{}
	err := LoadClientFS(myFS)
	if err != nil {
		return err
	}

	seq := &api.HistorySeq{}
	err = LoadHistorySeq(seq)
	if err != nil {
		if err == io.EOF {
			seq.Seq = 0
			seq.UploadSeq = 0
		} else {
			return err
		}
	}
	seqFile, err := os.OpenFile(".garage/history/historySeq", os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = seqFile.Close()
	}()
	seq.Seq = seq.Seq + 1
	b, err := proto.Marshal(seq)
	if err != nil {
		return err
	}
	_, err = seqFile.Write(b)
	if err != nil {
		return err
	}

	history := &api.SaveChanges{
		Seq:         seq.Seq,
		Msg:         msg,
		ChangeOrder: myFS.Changes,
	}

	// 수정/생성 된 파일을 다른 디렉터리로 이동(나중에 롤백할 수 있도록)
	err = MoveChangedFileAndSaveHistory(history)
	if err != nil {
		return err
	}

	return nil
}

func MoveChangedFileAndSaveHistory(saveInfo *api.SaveChanges) error {
	myFS := &api.ClientFS{}
	err := LoadClientFS(myFS)

	if err != nil {
		return err
	}
	savePath := ".garage/history/changeOrder_" + strconv.Itoa(int(saveInfo.Seq))
	err = os.MkdirAll(savePath, 0755)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(savePath+"/"+"save_"+strconv.Itoa(int(saveInfo.Seq)), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	b, err := proto.Marshal(saveInfo)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
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

	for k, v := range myFS.Changes {
		if v == "delete" {
			continue
		}
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

func ShowHistory() error {
	seq := &api.HistorySeq{}
	err := LoadHistorySeq(seq)
	if err != nil {
		if err == io.EOF {
			seq.Seq = 0
			seq.UploadSeq = 0
		} else {
			return err
		}
	}
	num := seq.Seq

	if seq.UploadSeq == 0 {
		if seq.Seq == 0 {
			fmt.Println("로컬의 저장된 변경사항이 없다.")
		}
		fmt.Println("리모트에 저장된 파일이 없다.")
	}
	for i := 1; i <= int(num); i++ {
		path := ".garage/history/changeOrder_" + strconv.Itoa(i) + "/save_" + strconv.Itoa(i)

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info := &api.SaveChanges{}

		proto.Unmarshal(b, info)

		fmt.Printf("save Num - %d\n", info.Seq)
		fmt.Printf("save Message - %s\n", info.Msg)
		fmt.Println("[change orders]")
		for k, v := range info.ChangeOrder {
			fmt.Printf("[%s : %s]\n", k, v)
		}
		fmt.Println()

		if i == int(seq.UploadSeq) {
			fmt.Println("----------------------------- complete upload to remote")
		}
	}

	return nil
}

func Push() error {
	seq := &api.HistorySeq{}
	err := LoadHistorySeq(seq)
	if err != nil {
		return err
	}
	if seq.Seq == seq.UploadSeq {
		return fmt.Errorf("모든 저장된 변경사항이 push 된 상태입니다")
	}
	saveSeq := seq.Seq
	upLoadSeq := seq.UploadSeq

	for {
		upLoadSeq = upLoadSeq + 1
		changeInfo := &api.SaveChanges{}
		err := LoadSaveChanges(changeInfo, int(upLoadSeq))
		if err != nil {
			return err
		}

		user := &api.UserInfo{}
		err = LoadUserInfo(user)
		if err != nil {
			return err
		}

		conn, garageClient, ctx, err := serverConn()
		if err != nil {
			return err
		}
		defer conn.Close()

		rootDir := fmt.Sprintf("./root/%s_%s", user.Id, user.GarageName)
		fmt.Println("서버로 푸쉬합니다.")
		var endFlag bool = false
		count := 1
		for k, v := range changeInfo.ChangeOrder {
			if count == len(changeInfo.ChangeOrder) {
				endFlag = true
			}
			path := fmt.Sprintf(".garage/history/changeOrder_%d/%s", upLoadSeq, k[2:])
			fileData := make([]byte, 4096)
			if v != "delete" {
				fd, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				fileData = fd
			}
			file := &api.File{
				RootDir:  rootDir,
				FilePath: k, // rootDir + "/" + i[2:]
				Desc:     v,
				FileData: fileData,
				EndFile:  endFlag,
			}

			response, err := garageClient.UploadFiles(ctx, file)
			if err != nil {
				return err
			}
			count++
			fmt.Println(response)

		}
		if upLoadSeq == saveSeq {
			break
		}
	}

	newSeq := &api.HistorySeq{
		Seq:       saveSeq,
		UploadSeq: upLoadSeq,
	}

	fd, err := os.OpenFile(".garage/history/historySeq", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()
	b, err := proto.Marshal(newSeq)
	if err != nil {
		return err
	}
	_, err = fd.Write(b)
	if err != nil {
		return err
	}
	return nil
}

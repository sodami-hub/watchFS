/*
db - garage
user - garage
pw - garagegarage


CREATE TABLE users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    id VARCHAR(100) NOT NULL UNIQUE,
    pw VARCHAR(100) NOT NULL
);

CREATE TABLE garages (
    garage_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    garage_name VARCHAR(100) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
*/

package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	api "github.com/sodami-hub/watchfs/api/v1"
)

// type User struct {
// 	id         string
// 	pw         string
// 	garageName string
// }

type GarageService struct {
	//user User

	mu sync.Mutex
	api.UnimplementedGarageServer
}

func databaseConn() (*sql.DB, error) {
	conn := "garage:garageservice@tcp(127.0.0.1:3306)/garage"

	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(3 * time.Minute)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)

	return db, nil
}

// garage join id pw
func (gs *GarageService) Join(_ context.Context, userInfo *api.UserInfo) (*api.Response, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	db, err := databaseConn()
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 접속 에러 : %v", err)
	}
	defer db.Close()

	insertQuery := "INSERT INTO users(id,pw) VALUES(?, ?)"
	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}
	res, err := stmt.Exec(userInfo.Id, userInfo.Pw)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return &api.Response{Message: "이미 존재하는 아이디입니다. 다른 아이디로 가입하세요."}, nil
		}
		return nil, err
	}
	lastId, err := res.LastInsertId() // 방금 추가된 데이터의 id를 가져온다.
	if err != nil {
		return nil, err
	}
	responseMsg := fmt.Sprintf("회원가입 완료. user_id : %d", lastId)
	return &api.Response{Message: responseMsg}, nil
}

// garage conn  => 로컬에 저장된 설정 파일의 정보로 서버 접속(id, pw, garage name)
func (gs *GarageService) LogIn(_ context.Context, userInfo *api.UserInfo) (*api.Response, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	db, err := databaseConn()
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 접속 에러 : %v", err)
	}
	defer db.Close()

	var pw string
	var user_id int
	// 결과에서 한줄만 가져온다.
	err = db.QueryRow("SELECT user_id,pw FROM users WHERE id=?", userInfo.Id).Scan(&user_id, &pw)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf("계정이 존재하지 않는다")
		}
		return nil, fmt.Errorf("error 106 line : %v", err)
	}
	if pw != userInfo.Pw {
		return nil, fmt.Errorf("패스워드를 다시 확인해주세요")
	}
	var garageName string
	rows, err := db.Query("SELECT garage_name FROM garages WHERE user_id=?", user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&garageName)
		if err != nil {
			return nil, err
		}
		if garageName == userInfo.GarageName {
			message := "login success / this garage name is " + garageName + " / 서비스를 시작합니다."
			return &api.Response{Message: message}, nil
		}
	}

	return &api.Response{Message: "사용자 계정 확인완료. garage를 생성해주세요. $ garage init [garagename]"}, nil
}

// garage init garagename => garagename으로 서비스 시작
func (gs *GarageService) InitGarage(_ context.Context, userInfo *api.UserInfo) (*api.Response, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	db, err := databaseConn()
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 접속 에러 : %v", err)
	}
	defer db.Close()

	var pw string
	var user_id int
	// 결과에서 한줄만 가져온다.
	err = db.QueryRow("SELECT user_id,pw FROM users WHERE id=?", userInfo.Id).Scan(&user_id, &pw)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf("존재하지 않는 계정입니다. 계정을 다시 등록해주세요")
		}
		return nil, fmt.Errorf("error 157 line : %v", err)
	}

	var garageName string
	rows, err := db.Query("SELECT garage_name FROM garages WHERE user_id=?", user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&garageName)
		if err != nil {
			return nil, err
		}
		if garageName == userInfo.GarageName {
			return nil, fmt.Errorf("당신의 계정에 이미 존재하는 garage 입니다. 다른 이름으로 설정하세요")
		}
	}

	insertQuery := "INSERT INTO garages(user_id,garage_name) VALUES(?, ?)"
	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}
	res, err := stmt.Exec(user_id, userInfo.GarageName)
	if err != nil {
		return nil, err
	}
	lastId, err := res.LastInsertId() // 방금 추가된 데이터의 id를 가져온다.
	if err != nil {
		return nil, err
	}

	dirPath := "./root/" + userInfo.Id + "_" + userInfo.GarageName
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return nil, err
	}
	msg := "당신의 계정으로 현재 디렉터리에 garage가 생성됐습니다. 서비스를 시작할 수 있습니다. garage_id :" + strconv.Itoa(int(lastId))
	return &api.Response{Message: msg}, nil
}

func (gs *GarageService) UploadFiles(_ context.Context, file *api.File) (*api.Response, error) {
	filePath := file.FilePath
	/*

		file에서 디렉터리만 생성한다.
		filepath.Dir 함수는 주어진 경로에서 디렉터리 부분을 반환한다.
		예를 들어, filepath.Dir("./temp/a")를 호출하면 "./temp/"를 반환합니다.
	*/

	dir := filepath.Dir(filePath)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}
	}

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	_, err = fd.WriteString(string(file.FileData))
	if err != nil {
		return nil, err
	}

	if file.EndFile {
		return &api.Response{Message: "upload complete"}, nil
	} else {
		return &api.Response{Message: "uploading..."}, nil
	}
}

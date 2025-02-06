/*
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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	api "github.com/sodami-hub/watchfs/api/v1"
)

type User struct {
	id         string
	pw         string
	garageName string
}

type GarageService struct {
	user User

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

// garage conn id pw
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
		return nil, err
	}
	lastId, err := res.LastInsertId() // 방금 추가된 데이터의 id를 가져온다.
	if err != nil {
		return nil, err
	}
	responseMsg := fmt.Sprintf("create new User. user_id : %d", lastId)
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
	var user_id string
	// 결과에서 한줄만 가져온다.
	err = db.QueryRow("SELECT user_id,pw FROM users WHERE id=?", userInfo.Id).Scan(&user_id, &pw)
	if err != nil {
		return nil, err
	}
	if pw == "" {
		return nil, fmt.Errorf("계정이 존재하지 않는다")
	}
	var garageName string
	rows, err := db.Query("SELECT garage_name FROM garages WHERE=?", user_id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&garageName)
		if err != nil {
			return nil, err
		}
		if garageName == userInfo.GarageName {
			message := "login success / this garage name is" + garageName
			return &api.Response{Message: message}, nil
		}
	}

	return nil, fmt.Errorf("현재 디렉터리에 생성된 garage가 없습니다. 생성해주세요. \n garage init [garagename]")
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
	var user_id string
	// 결과에서 한줄만 가져온다.
	err = db.QueryRow("SELECT user_id,pw FROM users WHERE id=?", userInfo.Id).Scan(&user_id, &pw)
	if err != nil {
		return nil, err
	}
	if pw == "" {
		return nil, fmt.Errorf("계정이 존재하지 않는다")
	}

	var garageName string
	rows, err := db.Query("SELECT garage_name FROM garages WHERE=?", user_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(&garageName)
		if err != nil {
			return nil, err
		}
		if garageName == userInfo.GarageName {
			return nil, fmt.Errorf("이미 존재하는 이름입니다. 다른 이름으로 설정하세요")
		}
	}

	dirPath := "./root/" + userInfo.Id + "_" + userInfo.GarageName
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return nil, err
	}
	return &api.Response{Message: "당신의 계정으로 현재 디렉터리에 garage가 생성됐습니다. 서비스를 시작합니다."}, nil
}

func (gs *GarageService) UploadFiles(_ context.Context, file *api.File) (*api.Response, error) {
	filePath := file.FilePath

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

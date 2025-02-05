/*
CREATE TABLE users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    id VARCHAR(100) NOT NULL,
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
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

func (gs *GarageService) Join(_ context.Context, userInfo *api.UserInfo) (*api.Response, error) {

	return &api.Response{Message: "ok"}, nil
}

func (gs *GarageService) Cert(_ context.Context, userInfo *api.UserInfo) (*api.Response, error) {

	return &api.Response{Message: "ok"}, nil
}

func (gs *GarageService) InitGarage(_ context.Context, userInfo *api.UserInfo) (*api.Response, error) {

	return &api.Response{Message: "ok"}, nil
}

func (gs *GarageService) UploadFiles(_ context.Context, file *api.File) (*api.Response, error) {

	return &api.Response{Message: "ok"}, nil
}

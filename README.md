# watchfs
매개변수로 받은 경로의 파일 시스템의 변화를 감시하는 라이브러리(새로운 파일 추가, 수정, 등...) 

## 구조체
```
type fs struct {
	initDir     string 
	directories []string
	allFiles    []map[string]string
	changes     []map[string]string
}
```
- initDir : root 디렉터리의 경로 저장
- directories : root 디렉터리 하위의 모든 디렉터리 경로를 저장하는 슬라이스
- allFiles : root 디렉터리하위의 모든 파일들의 경로(이름) 및 수정시간을 저장하는 슬라이스 맵
- changes : root 디렉터리 하위의 모든 경로안에서 변화를 감지하고 변화된 파일 또는 디렉터리의 정보를 저장하는 맵 슬라이스

## NewWatcher() 함수
- watcher 인스턴스 생성하는 함수

## DirSearch() 메서드 
- 구조체 fs의 메서드로 root 디렉터리 하위의 모든 파일시스템 정보를 가져온다. 
- 클라이언트의 초기 파일시스템 정보를 가져오는데 사용된다.
- 이후에는 이 정보를 DB에 저장하고 재접속시 DB의 정보와 현재 파일시스템의 차이를 비교해서 변경된 사항을 변경사항 목록에 표시되도록 한다.

## Watch() 메서드 
- root 디렉터리 하위의 모든 변경을 감시하고, 변경사항을 fs.changes에 저장한다.

### Wahch() 메서드의 세부적인 기능이다.
#### watchfs 패키지의 테스트를 위해서 만든 /main/main.go 를 실행한다. - 초기 루트 디렉터리 아래의 디렉터리 목록과 파일들 그리고 파일드의 최종 수정시간을 확인할 수 있다.
```
$ go run main.go ./
디렉토리 리스트
./
./temp
디렉토리 / 파일:수정시간 리스트
filename : ./main.go  //  modtime: 2025-01-31 12:45:37.687637885 +0900 KST
filename : ./temp/c  //  modtime: 2025-02-01 21:59:57.280189719 +0900 KST
filename : ./a  //  modtime: 2025-02-01 21:59:30.44578841 +0900 KST
filename : ./b  //  modtime: 2025-02-01 21:59:38.183893229 +0900 KST
filename : ./e  //  modtime: 2025-02-01 22:00:11.76844866 +0900 KST
변경 파일 목록
```
#### ./a 파일을 수정해 보겠다. 파일 수정 메시지와 함께 변경 파일 목록에 ./a의 최종 변경시간이 표시된다. 
```
파일 수정
디렉토리 리스트
./
./temp
디렉토리 / 파일:수정시간 리스트
filename : ./a  //  modtime: 2025-02-01 21:59:30.44578841 +0900 KST
filename : ./b  //  modtime: 2025-02-01 21:59:38.183893229 +0900 KST
filename : ./e  //  modtime: 2025-02-01 22:00:11.76844866 +0900 KST
filename : ./main.go  //  modtime: 2025-01-31 12:45:37.687637885 +0900 KST
filename : ./temp/c  //  modtime: 2025-02-01 21:59:57.280189719 +0900 KST
변경 파일 목록
filename : ./a // 변경사항 : 2025-02-01 22:02:58.194063191 +0900 KST
```
#### 파일 생성 ./d 파일을 생성했다. 변경사항으로는 create로 표시된다.
```
파일생성
디렉토리 리스트
./
./temp
디렉토리 / 파일:수정시간 리스트
filename : ./e  //  modtime: 2025-02-01 22:00:11.76844866 +0900 KST
filename : ./main.go  //  modtime: 2025-01-31 12:45:37.687637885 +0900 KST
filename : ./temp/c  //  modtime: 2025-02-01 21:59:57.280189719 +0900 KST
filename : ./a  //  modtime: 2025-02-01 21:59:30.44578841 +0900 KST
filename : ./b  //  modtime: 2025-02-01 21:59:38.183893229 +0900 KST
변경 파일 목록
filename : ./a // 변경사항 : 2025-02-01 22:02:58.194063191 +0900 KST
filename : ./d // 변경사항 : create

```
#### 디렉토리 생성 - ./temp2 디렉터리를 생성했다. 디렉토리 생성시에는 디렉토리 리스트에만 추가가 된다.
```
디렉터리 생성
디렉토리 리스트
./
./temp
./temp2
디렉토리 / 파일:수정시간 리스트
filename : ./a  //  modtime: 2025-02-01 21:59:30.44578841 +0900 KST
filename : ./b  //  modtime: 2025-02-01 21:59:38.183893229 +0900 KST
filename : ./e  //  modtime: 2025-02-01 22:00:11.76844866 +0900 KST
filename : ./main.go  //  modtime: 2025-01-31 12:45:37.687637885 +0900 KST
filename : ./temp/c  //  modtime: 2025-02-01 21:59:57.280189719 +0900 KST
변경 파일 목록
filename : ./a // 변경사항 : 2025-02-01 22:02:58.194063191 +0900 KST
filename : ./d // 변경사항 : create
```
#### 디렉터리 내부에 파일 생성 - ./temp2/abcd 파일을 추가해보겠다. ./temp2/abcd 가 create 상태로 추가됐다.
```
파일생성
디렉토리 리스트
./
./temp
./temp2
디렉토리 / 파일:수정시간 리스트
filename : ./e  //  modtime: 2025-02-01 22:00:11.76844866 +0900 KST
filename : ./main.go  //  modtime: 2025-01-31 12:45:37.687637885 +0900 KST
filename : ./temp/c  //  modtime: 2025-02-01 21:59:57.280189719 +0900 KST
filename : ./a  //  modtime: 2025-02-01 21:59:30.44578841 +0900 KST
filename : ./b  //  modtime: 2025-02-01 21:59:38.183893229 +0900 KST
변경 파일 목록
filename : ./d // 변경사항 : create
filename : temp2/abcd // 변경사항 : create
filename : ./a // 변경사항 : 2025-02-01 22:02:58.194063191 +0900 KST
```
#### 파일 삭제 - ./e 파일을 삭제한다. 파일 목록에 해당 파일이 포함돼 있지만, 변경 내용에 해당파일의 변경사항으로 delete가 표시되는 것을 확인할 수 있다. 
```
파일삭제 ./e
디렉토리 리스트
./
./temp
./temp2
디렉토리 / 파일:수정시간 리스트
filename : ./a  //  modtime: 2025-02-01 21:59:30.44578841 +0900 KST
filename : ./b  //  modtime: 2025-02-01 21:59:38.183893229 +0900 KST
filename : ./e  //  modtime: 2025-02-01 22:00:11.76844866 +0900 KST
filename : ./main.go  //  modtime: 2025-01-31 12:45:37.687637885 +0900 KST
filename : ./temp/c  //  modtime: 2025-02-01 21:59:57.280189719 +0900 KST
변경 파일 목록
filename : ./a // 변경사항 : 2025-02-01 22:02:58.194063191 +0900 KST
filename : ./d // 변경사항 : create
filename : temp2/abcd // 변경사항 : create
filename : ./e // 변경사항 : delete
```
#### 새롭게 생성된 파일을 삭제하는 경우에는 변경 파일 목록에서 해당 파일이 삭제된다. 

## Save() 메서드
- DirSearch() 메서드로 fs.allFiles 필드의 정보를 갱신하고, 해당 내용을 데이터베이스와 서버에 저장한다. fs.chages는 초기화 된다.

### 데이터베이스
- mySQL을 사용한다. 사용자정보(id/pw) 테이블, 사용자 레포지토리 정보 테이블, 레포지토리의 모든 경로(디렉토리, 파일) 와 최종 수정시간에 대한 데이터가 있다.
### 서버
- 사용자id/root 디렉터리/ 하위에 모든 정보가 사용자의 로컬과 동일한 정보가 담겨있다. 
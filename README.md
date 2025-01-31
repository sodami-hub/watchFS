# watcher
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

## Watcher() 메서드 
- root 디렉터리 하위의 모든 변경을 감시하고, 변경사항을 fs.changes에 저장한다.

## Save() 메서드
- DirSearch() 메서드로 fs.allFiles 필드의 정보를 갱신하고, 해당 내용을 데이터베이스와 서버에 저장한다. fs.chages는 초기화 된다.

### 데이터베이스
- mySQL을 사용한다. 사용자정보(id/pw) 테이블, 사용자 레포지토리 정보 테이블, 레포지토리의 모든 경로(디렉토리, 파일) 와 최종 수정시간에 대한 데이터가 있다.
### 서버
- 사용자id/root 디렉터리/ 하위에 모든 정보가 사용자의 로컬과 동일한 정보가 담겨있다. 
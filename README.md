# ipGachaProxy

이 프로젝트는 Go 언어를 사용하여 구현된 API 게이트웨이 기반의 IP 회전 프록시에 대해 설명합니다.

## 개요

- **언어**: Go
- **버전**: 1.22.1
- **지원 기능**
  - API 게이트웨이 디스커버리
  - 기본 프록시 기능

## 실행방법

```
go mod init github.com/MyungSub0519/ipGachaProxy
go mod tidy
go build -a -installsuffix cgo -o ipgachaproxy .
./ipgachaproxy
```

```
docker build . -t ipgachaproxy:0.0.1
docker run ipgachaproxy:0.0.1 -p 8510:8510
```

## 목표

### 1차 목표: 프로토타입

1. **기본 프록시 기능**

   - 요청을 포워딩하는 기본적인 프록시 기능 구현.

2. **API 게이트웨이**
   - 명령어로 생성되며, 생성된 항목들이 URL에 매핑되어 IP 회전을 담당합니다. (디스커버리)

### 2차 목표: 자동 생성

- API 게이트웨이의 항목들이 자동으로 생성되고 관리될 수 있는 기능 구현.

### 3차 목표: 프록시 기능 강화

- 프록시 기능의 성능과 보안 강화.

## 산출물

- **데몬(Daemon)**
  - 프로젝트의 실행 파일로, 백그라운드에서 지속적으로 서비스를 제공합니다.

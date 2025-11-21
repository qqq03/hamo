# Hamo - Jinju Museum RAG System

진주 박물관 RAG(Retrieval-Augmented Generation) 시스템 - Go 백엔드

## 기능
- AWS RDS MySQL 연동
- RAG 기반 문서 검색
- LLM 통합 준비
- AWS Secrets Manager 지원

## 환경 설정

### 로컬 개발 환경
`.env` 파일 생성:
```env
DB_USER=your_username
DB_PASS=your_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=museumdb
USE_SECRETS_MANAGER=false
```

### AWS EC2 프로덕션 환경

#### 1. AWS Secrets Manager에 DB 정보 저장
AWS Console → Secrets Manager → "Store a new secret"

Secret 타입: `Other type of secret`

Secret 값 (JSON 형식):
```json
{
  "username": "your_db_user",
  "password": "your_db_password",
  "host": "your-rds-endpoint.rds.amazonaws.com",
  "port": "3306",
  "dbname": "museumdb"
}
```

Secret 이름: `hamo/rds/credentials`

#### 2. EC2 IAM 역할 설정
EC2 인스턴스에 다음 권한이 있는 IAM 역할 연결:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:hamo/rds/credentials-*"
    }
  ]
}
```

#### 3. 환경 변수 설정
EC2에서 실행 시:
```bash
export USE_SECRETS_MANAGER=true
export SECRET_NAME=hamo/rds/credentials  # 기본값, 생략 가능
```

## 실행 방법

### 로컬에서 실행
```bash
go run main.go
```

### EC2에서 실행
```bash
# GitHub에서 코드 다운로드
git clone https://github.com/qqq03/hamo.git
cd hamo

# 의존성 설치
go mod download

# 환경 변수 설정
export USE_SECRETS_MANAGER=true

# 빌드 및 실행
go build -o hamo
./hamo
```

## API 엔드포인트

### 1. 데이터 조회
```
GET /api/data?id=1
```

### 2. RAG 질의
```
POST /api/rag
Content-Type: application/json

{
  "query": "진주 박물관에 대해 알려줘"
}
```

## 프로젝트 구조
```
hamo/
├── main.go           # 메인 애플리케이션
├── go.mod            # Go 모듈 정의
├── .env              # 로컬 환경 변수 (Git 제외)
├── .gitignore        # Git 무시 파일
└── README.md         # 프로젝트 문서
```

## 보안 주의사항
- `.env` 파일은 절대 Git에 업로드하지 마세요
- 프로덕션에서는 반드시 AWS Secrets Manager 사용
- SSH 키(.pem, .ppk)는 Git에 업로드하지 마세요

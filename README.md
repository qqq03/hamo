# Hamo - 진주 박물관 RAG 시스템

진주 박물관 RAG(Retrieval-Augmented Generation) 시스템 Go 백엔드 서버

## 📋 개요

이 프로젝트는 진주 박물관 데이터를 기반으로 한 RAG 시스템입니다. AWS RDS MySQL 데이터베이스와 연동하여 박물관 전시물 정보를 조회하고, 향후 LLM API와 통합하여 지능형 질의응답 서비스를 제공할 예정입니다.

### 주요 기능
- ✅ AWS RDS MySQL 데이터베이스 연동
- ✅ RESTful API (데이터 조회, RAG 엔드포인트)
- ✅ AWS Secrets Manager 통합 (보안 강화)
- ✅ 로컬/프로덕션 환경 분리
- 🚧 LLM API 통합 (예정)
- 🚧 벡터 검색 (OpenSearch/ChromaDB) (예정)

## 🏗️ 아키텍처

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   Client    │─────▶│   Go Server  │─────▶│  AWS RDS    │
│  (API 요청)  │      │   (8080)     │      │   MySQL     │
└─────────────┘      └──────────────┘      └─────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │AWS Secrets   │
                     │Manager       │
                     └──────────────┘
```

### 계층 구조
- **Handler**: HTTP 요청/응답 처리
- **Service**: 비즈니스 로직 (RAG 처리)
- **Repository**: 데이터베이스 접근
- **Model**: 데이터 구조 정의

## 🚀 빠른 시작

### 사전 요구사항
- Go 1.21 이상
- MySQL 8.0 이상 (또는 AWS RDS)
- (선택) AWS 계정 (Secrets Manager 사용 시)

### 로컬 개발 환경 설정

1. **저장소 클론**
```bash
git clone https://github.com/qqq03/hamo.git
cd hamo
```

2. **환경 변수 설정**
```bash
# .env.example을 복사하여 .env 생성
cp .env.example .env
```

`.env` 파일 수정:
```env
DB_USER=your_username
DB_PASS=your_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=museumdb
USE_SECRETS_MANAGER=false
```

3. **의존성 설치 및 실행**
```bash
go mod download
go run main.go
```

서버가 `http://localhost:8080`에서 실행됩니다.

## ☁️ AWS 프로덕션 배포

### 1. AWS Secrets Manager 설정

**Secret 생성:**
- AWS Console → Secrets Manager → "Store a new secret"
- Secret type: `Other type of secret`
- Secret 값 (JSON):
```json
{
  "username": "admin",
  "password": "your_secure_password"
}
```
- Secret name: `hamo/rds/credentials`

### 2. EC2 IAM 역할 설정

EC2 인스턴스에 다음 정책 추가:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["secretsmanager:GetSecretValue"],
      "Resource": "arn:aws:secretsmanager:ap-northeast-2:*:secret:hamo/rds/credentials-*"
    }
  ]
}
```

### 3. EC2에서 애플리케이션 실행

```bash
# 저장소 클론
git clone https://github.com/qqq03/hamo.git
cd hamo

# 의존성 설치
go mod download

# 환경 변수 설정
export USE_SECRETS_MANAGER=true
export DB_HOST=your-rds-endpoint.rds.amazonaws.com
export DB_NAME=museumdb
export AWS_REGION=ap-northeast-2

# 빌드 및 실행
go build -o hamo
./hamo
```

**백그라운드 실행 (systemd 서비스):**
```bash
sudo nano /etc/systemd/system/hamo.service
```

```ini
[Unit]
Description=Hamo RAG Server
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/hamo
Environment="USE_SECRETS_MANAGER=true"
Environment="DB_HOST=your-rds.rds.amazonaws.com"
Environment="DB_NAME=museumdb"
Environment="AWS_REGION=ap-northeast-2"
ExecStart=/home/ubuntu/hamo/hamo
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable hamo
sudo systemctl start hamo
```

## 📡 API 사용법

서버는 포트 **8080**에서 실행되며, 두 개의 RESTful API 엔드포인트를 제공합니다.

### API 엔드포인트 목록

| 메서드 | 엔드포인트 | 설명 |
|--------|-----------|------|
| `GET` | `/api/data?id={item_seq}` | 특정 아이템 조회 |
| `POST` | `/api/rag` | RAG 기반 질의응답 |

---

### 1. 아이템 조회 API

**엔드포인트:** `GET /api/data`

**설명:** Item 테이블에서 특정 ITEM_SEQ의 박물관 전시물 정보를 조회합니다.

**쿼리 파라미터:**
- `id` 또는 `item_seq` (필수): 조회할 아이템의 시퀀스 번호

**요청 예시:**
```bash
# curl 사용
curl http://localhost:8080/api/data?id=1

# PowerShell 사용
Invoke-WebRequest -Uri "http://localhost:8080/api/data?id=1" -Method GET | Select-Object -Expand Content
```

**성공 응답 (200 OK):**
```json
{
  "theme_id": "TH001",
  "item_seq": 1,
  "item_name": "진주성",
  "item_desc": "임진왜란 당시 중요한 전투지",
  "latitude": 35.1920,
  "longitude": 128.0844,
  "target_age": 10
}
```

**에러 응답:**
- `400 Bad Request`: `id` 파라미터 누락
  ```
  item_seq 파라미터가 필요합니다 (예: ?id=1)
  ```
- `404 Not Found`: 해당 ID의 문서가 없음
  ```
  문서를 찾을 수 없습니다
  ```
- `500 Internal Server Error`: DB 조회 오류
  ```
  데이터 조회 중 서버 오류 발생
  ```

---

### 2. RAG 질의응답 API

**엔드포인트:** `POST /api/rag`

**설명:** RAG(Retrieval-Augmented Generation) 시스템을 통해 사용자 질문에 답변합니다. 현재는 DB에서 ID 1, 2번 문서를 기반으로 응답하며, 향후 벡터 검색 및 LLM API가 통합될 예정입니다.

**요청 헤더:**
```
Content-Type: application/json
```

**요청 본문:**
```json
{
  "query": "사용자 질문 내용"
}
```

**요청 예시:**
```bash
# curl 사용
curl -X POST http://localhost:8080/api/rag \
  -H "Content-Type: application/json" \
  -d '{"query":"진주 박물관에 대해 알려줘"}'

# PowerShell 사용
$body = @{query="진주 박물관에 대해 알려줘"} | ConvertTo-Json
Invoke-WebRequest -Uri "http://localhost:8080/api/rag" `
  -Method POST `
  -Body $body `
  -ContentType "application/json" | Select-Object -Expand Content
```

**성공 응답 (200 OK):**
```json
{
  "answer": "LLM 응답: 당신의 질문 '진주 박물관에 대해 알려줘'은(는) [진주성] 정보를 바탕으로 처리되었습니다.",
  "sources": ["진주성", "진주 박물관"]
}
```

**응답 필드 설명:**
- `answer` (string): LLM이 생성한 답변 (현재는 Placeholder)
- `sources` (array): RAG에 사용된 문서 출처 목록

**에러 응답:**
- `400 Bad Request`: 잘못된 JSON 형식 또는 query 누락
  ```
  요청 본문(JSON) 파싱 오류
  질문(Query) 내용이 비어있습니다
  ```
- `405 Method Not Allowed`: GET 등 다른 메서드 사용
  ```
  POST 메서드만 허용됩니다
  ```
- `500 Internal Server Error`: RAG 처리 중 오류
  ```
  LLM 처리 중 서버 오류 발생
  ```

---

### 테스트 도구

**브라우저에서 테스트:**
- 데이터 조회: http://localhost:8080/api/data?id=1

**Postman/Thunder Client 사용:**
1. 새 요청 생성
2. 메서드: `POST`
3. URL: `http://localhost:8080/api/rag`
4. Headers: `Content-Type: application/json`
5. Body (raw JSON):
   ```json
   {
     "query": "진주 박물관에 대해 알려줘"
   }
   ```

## 📂 프로젝트 구조

```
hamo/
├── main.go              # 메인 애플리케이션 코드
├── go.mod               # Go 모듈 정의
├── go.sum               # 의존성 체크섬
├── .env                 # 로컬 환경 변수 (Git 제외)
├── .env.example         # 환경 변수 템플릿
├── .gitignore           # Git 무시 파일
└── README.md            # 프로젝트 문서
```

## 🔧 환경 변수

| 변수명 | 필수 | 기본값 | 설명 |
|--------|------|--------|------|
| `USE_SECRETS_MANAGER` | ❌ | `false` | Secrets Manager 사용 여부 |
| `SECRET_NAME` | ❌ | `hamo/rds/credentials` | Secret 이름 |
| `AWS_REGION` | ❌ | `ap-northeast-2` | AWS 리전 |
| `DB_HOST` | ✅ | - | 데이터베이스 호스트 |
| `DB_PORT` | ❌ | `3306` | 데이터베이스 포트 |
| `DB_NAME` | ✅ | - | 데이터베이스 이름 |
| `DB_USER` | ✅* | - | DB 사용자명 (로컬 전용) |
| `DB_PASS` | ✅* | - | DB 비밀번호 (로컬 전용) |
| `SKIP_DB_CHECK` | ❌ | `false` | DB 연결 확인 건너뛰기 |

*: 로컬 개발 시 필수, 프로덕션에서는 Secrets Manager 사용

## 🔒 보안 주의사항

### ⚠️ 절대 하지 말아야 할 것
- `.env` 파일을 Git에 커밋
- 데이터베이스 비밀번호를 코드에 하드코딩
- SSH 키(.pem, .ppk)를 Git에 업로드
- Private 저장소라도 민감한 정보 업로드

### ✅ 권장 사항
- 프로덕션: AWS Secrets Manager 사용
- 로컬: `.env` 파일 사용 (`.gitignore`에 포함됨)
- IAM 역할 기반 인증 사용 (EC2 인스턴스)
- 정기적인 비밀번호 로테이션

## 🐛 트러블슈팅

**Q: DB 연결 실패 (connection refused)**
- RDS 보안 그룹에서 3306 포트 허용 확인
- VPC 설정 확인 (EC2와 RDS가 같은 VPC에 있어야 함)

**Q: Secrets Manager 조회 실패**
- EC2 IAM 역할에 `secretsmanager:GetSecretValue` 권한 확인
- Secret 이름과 리전이 올바른지 확인

**Q: 포트 8080이 이미 사용 중**
```bash
# 프로세스 확인 및 종료
lsof -i :8080
kill -9 <PID>
```

## 📝 TODO
- [ ] LLM API (Gemini/OpenAI) 통합
- [ ] 벡터 데이터베이스 (OpenSearch) 연동
- [ ] 임베딩 생성 및 유사도 검색
- [ ] 사용자 인증/인가
- [ ] 로깅 및 모니터링
- [ ] Docker 컨테이너화

## 📄 라이선스
이 프로젝트는 진주 박물관 프로젝트의 일부입니다.

## 👥 기여자
- [@qqq03](https://github.com/qqq03) - 초기 개발

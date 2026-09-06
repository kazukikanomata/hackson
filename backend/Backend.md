## 概要

ハッカソン用バックエンドの環境構築メモ。

| 項目           | 採用技術                               |
| -------------- | -------------------------------------- |
| 言語           | Go                                     |
| HTTPサーバ     | 標準ライブラリ `net/http`              |
| DB             | PostgreSQL 16（Docker Compose で起動） |
| 待ち受けポート | 8080                                   |

Caddy が `/api/*` のリクエストをこのサーバ（8080）に転送します。
フロントエンドからは `/api/hello` のような相対パスで到達します。

---

## セットアップ

```bash
brew install go
```

モジュールは初期化済みです（`go.mod`）。

```
module hackson/backend

go 1.24.3
```

外部ライブラリを追加する場合:

```bash
go get github.com/example/lib   # 追加
go mod tidy                     # 不要な依存の整理
```

> GitHub にプッシュする際は、モジュール名を
> `github.com/<ユーザ名>/hackson/backend` に変更しても構いません
> （`go.mod` の1行目を書き換えるだけです）。

---

## 起動

```bash
cd backend
go run main.go     # => Go server starting on :8080...
```

フロント込みで通しで動かす手順はリポジトリルートの `setup.md` を参照。

### 開発時によく使うコマンド

```bash
go run main.go     # 起動
go build ./...     # ビルド確認
go vet ./...       # 静的解析
go test ./...      # テスト
gofmt -l .         # 未整形ファイルの検出（何も出なければOK）
gofmt -w .         # 整形を適用
```

**コミット前に `gofmt -w .` をかけてください。** Go は整形スタイルが
言語標準で決まっているため、揃えておかないと差分が読みにくくなります。

### 動作確認

```bash
curl http://localhost:8080/api/hello    # => {"message": "Hello, Go API!!"}
curl http://localhost:8080/api/health   # => {"status": "ok"}
```

---

## エンドポイント

| パス              | 用途               | レスポンス                       |
| ----------------- | ------------------ | -------------------------------- |
| `GET /api/health` | ヘルスチェック     | `{"status": "ok"}`               |
| `GET /api/hello`  | 疎通確認用サンプル | `{"message": "Hello, Go API!!"}` |

### ハンドラの追加方法

`main.go` の `main()` 内に `http.HandleFunc` を追加します。

```go
http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"users": []}`)
})
```

**パスは必ず `/api/` から始めてください。** Caddy は `/api/*` だけを
このサーバに転送し、それ以外はフロントエンド（3000）に流します。

規模が大きくなったらハンドラをファイル分割し、必要ならルータ
（`chi` など）の導入を検討します。

---

## ヘルスチェックについて

`/api/health` は Docker Compose のコンテナ死活監視から叩かれます。

```yaml
healthcheck:
  test: ["CMD", "wget", "--spider", "-q", "http://localhost/api/health"]
  interval: 5s
  timeout: 3s
  retries: 5
  start_period: 5s
```

このエンドポイントが `unhealthy` になると、`depends_on` で待っている
Caddy が起動しません。**認証を挟んだり重い処理を入れたりしないでください。**

DBとの接続確認まで含めたい場合は `db.PingContext()` の結果を返す形に拡張します。

---

## DB接続（今後）

接続情報はリポジトリルートの `.env` で管理しています（**git管理外**）。

| 変数 | 用途 |
| --- | --- |
| `POSTGRES_USER` | DBユーザ名 |
| `POSTGRES_PASSWORD` | DBパスワード |
| `POSTGRES_DB` | DB名 |

| 項目 | 値 |
| --- | --- |
| ホスト | `localhost`（ネイティブ実行） / `db`（コンテナ内） |
| ポート | 5432 |

### Goから読む場合

**ソースに直書きせず、必ず環境変数から読んでください。**

```go
dsn := fmt.Sprintf(
	"postgres://%s:%s@%s:5432/%s?sslmode=disable",
	os.Getenv("POSTGRES_USER"),
	os.Getenv("POSTGRES_PASSWORD"),
	os.Getenv("DB_HOST"),
	os.Getenv("POSTGRES_DB"),
)
```

ネイティブ実行時は `.env` が自動で読まれないため、以下のいずれかで渡します。

```bash
# 方法1: その場で読み込む
set -a && source ../.env && set +a && go run main.go

# 方法2: godotenv を使う
go get github.com/joho/godotenv
```

ドライバは `pgx` が標準的です（`go get github.com/jackc/pgx/v5`）。

---

## 注意点

- compose の `backend` サービスはまだ **nginx のダミー**です。
  実物に差し替える際は Dockerfile が必要になります。
- Go を `scratch` / `distroless` でビルドするとシェルが無いため、
  `CMD-SHELL` 形式のヘルスチェックが使えません。`alpine` ベースにするか、
  ヘルスチェック用の小さなバイナリを同梱してください。

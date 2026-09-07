## 概要

ハッカソン用バックエンドの環境構築メモ。

| 項目           | 採用技術                               |
| -------------- | -------------------------------------- |
| 言語           | Go                                     |
| HTTPサーバ     | 標準ライブラリ `net/http`              |
| DB             | PostgreSQL 16（Docker Compose で起動） |
| 待ち受けポート | 8080                                   |

Caddy が `/api/*` のリクエストをこのサーバ（8080）に転送します。フロントエンドからは `/api/hello` のような相対パスで到達します。

---

## 構成

Goのアプリケーションコードにおいて、ソースコードを更新するたびにソースをビルドし直して実行するという手間を省くため、ソースコードを更新し、保存したら自動でビルドと実行を行うために、airを用いる。

## セットアップ

**Docker で動かす場合、Go のインストールは不要です。** コンテナ内の Go がビルドします。
ネイティブで動かしたい場合のみインストールしてください。

```bash
brew install go
```

### Goのバージョン管理

Node の `.nvmrc` + nvm に相当する仕組みは、**Go に標準で組み込まれています**。
別途ツールを入れる必要はありません。`go.mod` の `toolchain` がそれです。

```
module hackson/backend

go 1.24.3          ← 必要な最低バージョン
toolchain go1.26.8 ← 実際に使うバージョン
```

Go 1.21 以降は `GOTOOLCHAIN=auto`（既定値）が有効で、手元の Go が
`toolchain` の指定より古い場合、**その版を自動でダウンロードして使います**。

```bash
cd backend && go version    # => go version go1.26.8 ...
```

手元の Go が新しい場合はダウングレードせず、そのまま新しい方が使われます。
つまり **Go さえ入っていれば、チーム内でバージョンを手動で揃える必要はありません**。

> `toolchain` の版を上げるときは `go get go@1.27.0` のように実行するか、
> `go.mod` を直接編集します。Docker 側の `golang:1.26-alpine` も
> 合わせて更新してください。

### 外部ライブラリの追加

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

### 自動リロード（air）

`go run main.go` は起動時に1回コンパイルするだけなので、**コードを変更しても
再起動しないと反映されません**。`air` を使うと保存を検知して自動で再ビルド・
再起動します。設定は `.air.toml` にあります。

**方法1: Docker で動かす（Goのインストール不要）**

```bash
docker compose up -d db backend
docker compose logs -f backend    # ビルドログを見る
```

ホストの `backend/` をコンテナにマウントしているので、**イメージの再ビルドは不要**です。
保存すると約1秒で反映されます。8080番はホストに公開しているため、
フロントエンドをネイティブで動かしていてもそのまま繋がります。

**方法2: ネイティブで動かす**

```bash
go install github.com/air-verse/air@latest   # 初回のみ
cd backend && air                            # go run main.go の代わり
```

> `air` は `$(go env GOPATH)/bin` に入ります。`air: command not found` になる場合は
> PATH を通してください。
> ```bash
> export PATH="$PATH:$(go env GOPATH)/bin"
> ```
> なお `air` は Go 1.26 以上を要求しますが、`go.mod` の `toolchain go1.26.8` により
> 自動で満たされるため、手元の Go を手動で更新する必要はありません。

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

---

このエンドポイントが `unhealthy` になると、`depends_on` で待っている
Caddy が起動しません。**認証を挟んだり重い処理を入れたりしないでください。**

DBとの接続確認まで含めたい場合は `db.PingContext()` の結果を返す形に拡張します。

---

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

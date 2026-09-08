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

## 原則

Goのアプリケーションロジックは原則、標準ライブラリを用いる。外部ライブラリを使う時は例外的に検討する。

### 例外: DBドライバ

**Goの標準ライブラリにPostgresドライバは存在しません。** `database/sql` は
ドライバを差し込むための共通インターフェースであり、実体は必ず外部パッケージになります。

そのため以下の形で、外部依存をドライバ1つに封じ込めています。

```go
import (
	"database/sql"                     // クエリ・トランザクションはこの標準APIで書く
	_ "github.com/jackc/pgx/v5/stdlib" // ドライバ登録のみ。コードから直接呼ばない
)

db, err := sql.Open("pgx", dsn)
```

`_` import によりドライバは登録されるだけで、アプリケーションロジックからは
一切参照しません。将来ドライバを差し替える場合も、この2行の変更で済みます。

> `pgxpool` など pgx 独自のAPIを直接使うと、ロジックが特定ライブラリに
> 密結合するため採用していません。

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
>
> ```bash
> export PATH="$PATH:$(go env GOPATH)/bin"
> ```
>
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

| パス              | 用途                     | レスポンス                              |
| ----------------- | ------------------------ | --------------------------------------- |
| `GET /api/health` | ヘルスチェック（DB疎通） | 200 `{"status": "ok"}`                  |
|                   | DB接続失敗時             | 500 `{"status": "db connection failed"}` |
| `GET /api/hello`  | 疎通確認用サンプル       | 200 `{"message": "Hello, Go API!!"}`    |

`/api/health` は `db.PingContext()` の結果を返します。DBが停止していると500になり、
compose の healthcheck が `unhealthy` を検知します。DBが復帰すればプールが自動で
接続を張り直すため、**アプリの再起動は不要**です。

### ハンドラの追加方法

`main.go` の `main()` 内に `http.HandleFunc` を追加します。

```go
http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"users": []}`)
})
```

**パスは必ず `/api/` から始めてください。** プロキシは `/api/*` だけを
このサーバに転送し、それ以外はフロントエンド（3000）に流します。

---

## ヘルスチェックについて

`/api/health` は Docker Compose のコンテナ死活監視から叩かれます。

```yaml
healthcheck:
  test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/api/health"]
  interval: 5s
  timeout: 3s
  retries: 10
  start_period: 120s
```

このエンドポイントが `unhealthy` になると、`depends_on` で待っている
Caddy が起動しません。**認証を挟んだり重い処理を入れたりしないでください。**

---

## DB接続

`connection.go` の `InitDB()` で接続します。実装済みです。

### 構成

```go
import (
	"database/sql"                     // クエリ・トランザクションはこの標準APIで書く
	_ "github.com/jackc/pgx/v5/stdlib" // ドライバ登録のみ。直接呼ばない
)

dbpool, err := sql.Open("pgx", dsn)
```

- `*sql.DB` は**起動時に1回だけ**作り、ハンドラで使い回す（リクエスト毎に作らない）
- `defer dbpool.Close()` は `main` に置く（`InitDB` 内に置くと返す前に閉じてしまう）
- 接続失敗時は `log.Fatal` で落とす（黙って起動させない）
- DB復帰時はプールが自動で接続を張り直すため、アプリの再起動は不要

### 接続情報

リポジトリルートの `.env` で管理しています（**git管理外**）。

| 変数 | 用途 |
| --- | --- |
| `POSTGRES_USER` | DBユーザ名 |
| `POSTGRES_PASSWORD` | DBパスワード |
| `POSTGRES_DB` | DB名 |
| `DB_HOST` | 接続先ホスト（compose が `db` を渡す） |

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

> `?sslmode=disable` の `?` は**1つ**です。2つ書くと `?sslmode` という未知の
> パラメータとして扱われ、`FATAL: unrecognized configuration parameter` で
> 起動できません（エラーメッセージが分かりにくいので注意）。

ネイティブ実行時は `.env` が自動で読まれないため、以下のいずれかで渡します。

```bash
# 方法1: その場で読み込む
set -a && source ../.env && set +a && go run main.go

# 方法2: godotenv を使う
go get github.com/joho/godotenv
```

---

## 注意点

- compose の `backend` は本物の Go に差し替え済みですが、**開発専用の構成**です
  （`golang:1.26-alpine` + バインドマウント + air）。本番デプロイには別途
  マルチステージビルドの Dockerfile が必要です。
- Go を `scratch` / `distroless` でビルドするとシェルが無いため、
  `CMD-SHELL` 形式のヘルスチェックが使えません。`alpine` ベースにするか、
  ヘルスチェック用の小さなバイナリを同梱してください。
- テーブル定義（マイグレーション）の方針は未決です。

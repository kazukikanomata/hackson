# セットアップ手順

初めての人は [README.md](README.md) のクイックスタートから読んでください。
このファイルは詳細な手順と、構成の背景を説明します。

## 構成

| レイヤ | 技術 | ポート | 普段の起動方法 |
| --- | --- | --- | --- |
| フロントエンド | Vite + React + TypeScript + Tailwind CSS v4 + shadcn/ui | 3000 | ネイティブ（`pnpm dev`） |
| バックエンド | Go (`net/http`) + air | 8080 | Docker |
| DB | PostgreSQL 16 | 5432 | Docker |
| リバースプロキシ | Caddy | 80 / 443 | 普段は使わない |

普段の開発では **Caddy は不要**です。Vite の devProxy が `/api/*` を Go(8080) に
転送するため、`http://localhost:3000` だけで完結します。
Caddy は本番に近い構成を確認したいときだけ使います。

---

## 1. 必要なツール

| ツール | 用途 | 必須か |
| --- | --- | --- |
| Docker Desktop | DB とバックエンドの起動 | **必須** |
| nvm | Node のバージョン管理 | **必須** |
| Go | バックエンドをネイティブで動かす場合のみ | 任意 |
| Caddy | 構成全体を確認する場合のみ | 任意 |

```bash
# 必須
brew install --cask docker
brew install nvm

# 任意
brew install go
brew install caddy
```

バックエンドを Docker で動かす場合、**Go のインストールは不要**です。
コンテナ内の Go がビルドします。

### Node のバージョン

`frontend/.nvmrc` で固定しています。

```bash
cd frontend
nvm install && nvm use     # .nvmrc の版を入れて切り替え
corepack enable pnpm       # pnpm を有効化
```

`package.json` の `packageManager` で `pnpm@10.33.0` に固定しています。
**npm / yarn は使わないでください**（lockfile が壊れます）。

### Go のバージョン

`backend/go.mod` の `toolchain` で固定しています。**nvm のような別ツールは不要**です。

```
go 1.24.3          ← 必要な最低バージョン
toolchain go1.26.8 ← 実際に使うバージョン
```

Go 1.21 以降は `GOTOOLCHAIN=auto`（既定値）が有効で、手元の Go が古ければ
**`toolchain` に書かれた版を自動でダウンロードして使います**。
つまり Go さえ入っていれば、バージョンを手動で合わせる必要はありません。

```bash
cd backend && go version    # => go version go1.26.8 ...
```

> 手元の Go の方が新しい場合はダウングレードせず、そのまま新しい方が使われます。

---

## 2. 環境変数の準備

DBの認証情報は `.env` から読み込みます。**`.env` は git 管理外**なので、
clone 後に雛形からコピーしてください。

```bash
cp .env.example .env
```

| 変数 | 用途 |
| --- | --- |
| `POSTGRES_USER` | DBユーザ名 |
| `POSTGRES_PASSWORD` | DBパスワード |
| `POSTGRES_DB` | DB名 |

`.env` が無い、または値が空のまま `docker compose up` すると、
`required variable POSTGRES_USER is missing a value` というエラーで停止します
（黙って起動して後で困らないよう、意図的にそうしています）。

> 項目を追加したら **`.env.example` にもキーだけ追記**してください。
> そうしないと他のメンバーが何を設定すべきか分かりません。

---

## 3. フロントエンドの依存インストール

```bash
cd frontend
pnpm install
```

---

## 4. 起動

### パターンA: 普段の開発（推奨）

**DB とバックエンドを Docker、フロントエンドをネイティブ**で動かします。

```bash
# 1. DB + バックエンド（コード変更は自動反映される）
docker compose up -d db backend

# 2枚目のターミナル: フロントエンド
cd frontend && pnpm dev
```

ブラウザで **http://localhost:3000** を開く。
「Call Go API」ボタンで `Hello, Go API!!` が表示されれば疎通OK。

バックエンドは `air` がホストの `backend/` を監視しているので、
**`.go` を保存すると約1秒で自動的に再ビルド・再起動されます**。
ビルドエラーはログに出ます。

```bash
docker compose logs -f backend
```

> 初回だけ `air` の取得に1〜2分かかります。2回目以降はキャッシュが効いて
> 10秒程度で起動します。

フロントだけ触る日は `pnpm dev` のみでも構いません（API呼び出しは失敗しますが画面は出ます）。

### パターンB: バックエンドもネイティブ

Go のデバッガを使いたい場合など。

```bash
docker compose up -d db          # DBのみ

go install github.com/air-verse/air@latest   # 初回のみ
cd backend && air                            # 自動リロードあり
# または
cd backend && go run main.go                 # 自動リロードなし
```

> `air` は `$(go env GOPATH)/bin` に入ります。PATH が通っていない場合は
> シェルの設定に追記してください。
> ```bash
> export PATH="$PATH:$(go env GOPATH)/bin"
> ```

### パターンC: 構成全体の確認（Caddy込み）

Caddyfile やプロキシ設定を変更した後の検証用です。

```bash
docker compose up -d      # 全サービス
```

**http://localhost** で確認。終わったら `docker compose down`。

> ネイティブのCaddyとDockerのCaddyは同じ80番を使うため**同時に起動できません**。
> 切り替える際は `caddy stop` / `docker compose down` で必ず片方を止めてください。

### 停止

```bash
docker compose down          # コンテナ削除（DBのデータも消えます）
docker compose stop          # 停止のみ（データは残る）
```

> compose に volume を定義していないため、`down` すると**DBのデータは消えます**。
> 永続化が必要になったら `db` サービスに volume を追加してください。

---

## 5. Docker Compose とヘルスチェック

### 起動順の制御

`depends_on` に `condition: service_healthy` を付けているので、
**db が healthy になってから backend、両方が healthy になってから caddy** の順で起動します。
「DBはまだ起動中なのにアプリが繋ぎにいって落ちる」という事故が防げます。

```bash
docker compose ps          # STATUS 列が (healthy) になればOK
```

### 各サービスのチェック内容

| サービス | 中身 | チェック方法 |
| --- | --- | --- |
| db | PostgreSQL 16 | `pg_isready` |
| backend | Go + air（ホストの `backend/` をマウント） | `wget --spider http://localhost:8080/api/health` |
| frontend | **nginx のダミー** | `wget --spider http://localhost/` |
| caddy | Caddy | `wget --spider http://localhost:2019/config/`（管理API） |

パラメータの意味:

- `interval` — チェックの実行間隔
- `timeout` — 1回のチェックのタイムアウト
- `retries` — 連続何回失敗したら `unhealthy` にするか
- `start_period` — 起動直後の猶予期間。この間の失敗は `retries` にカウントされない

### 個別に状態を見る

```bash
# 状態と連続失敗回数
docker inspect --format '{{.State.Health.Status}} / {{.State.Health.FailingStreak}}' LB

# 直近のチェック結果（失敗理由の調査に使う）
docker inspect --format '{{json .State.Health.Log}}' LB | jq
```

### backend コンテナの仕組み

イメージ再ビルドなしでコード変更が反映されるのは、ホストのソースを
バインドマウントしているためです。

```yaml
volumes:
  - ./backend:/app                        # ホストのソースを直接参照
  - go-path:/go                           # 依存と air をキャッシュ
  - go-build-cache:/root/.cache/go-build  # ビルドキャッシュ
```

`go-path` と `go-build-cache` が無いと毎回フルビルドになり、起動が非常に遅くなります。

---

## 6. Caddyfile について

ネイティブ起動とDocker起動の両方で同じ `Caddyfile` を使えるよう、環境変数で上書きできるようにしています。

```
{$SITE_ADDRESS:localhost} {
    handle /api/* { reverse_proxy {$BACKEND_UPSTREAM:localhost:8080} }
    handle        { reverse_proxy {$FRONTEND_UPSTREAM:localhost:3000} }
}
```

| | SITE_ADDRESS | BACKEND_UPSTREAM | FRONTEND_UPSTREAM |
| --- | --- | --- | --- |
| ネイティブ（デフォルト） | `localhost` | `localhost:8080` | `localhost:3000` |
| Docker（compose で上書き） | `:80` | `backend:8080` | `frontend:80` |

**なぜ必要か**: コンテナ内の `localhost` はそのコンテナ自身を指すため、Docker では
`localhost:8080` ではなくサービス名 `backend` で名前解決する必要があります。
また `localhost { }` と書くとCaddyが自動HTTPS化して443にリダイレクトしますが、
composeでは80番しか公開していないため `:80` を指定しています。

---

## トラブルシューティング

**ポート3000が埋まっていて `pnpm dev` が起動しない**

`strictPort: true` を設定しているため、別ポートにずれずエラーで停止します。
先客を止めてください。

```bash
lsof -ti TCP:3000 -sTCP:LISTEN | xargs kill
```

**`http://localhost` が意図しない内容を返す**

過去に起動した Caddy が残って 80/443 を掴んでいる可能性があります。

```bash
caddy stop
docker compose down
lsof -nP -iTCP:80,443 -sTCP:LISTEN    # 残っていれば kill <PID>
```

**Goのコードを変えたのに反映されない**

`go run main.go` で起動している場合、自動反映されません（起動時に1回コンパイルするだけ）。
`Ctrl+C` で止めて再実行するか、`air` を使ってください。

**backend コンテナが unhealthy から戻らない**

ビルドエラーの可能性が高いです。ログを確認してください。

```bash
docker compose logs backend
```

---

## 補足 / TODO

- compose の `frontend` サービスはまだ **nginx のダミー**です（`backend` は実物に差し替え済み）。
  フロントもDocker化する場合は、`backend` と同様にバインドマウント + `pnpm dev --host` の構成にします。
- compose に volume が無いため、`docker compose down` でDBのデータが消えます。
- OpenAPI からの型自動生成（`oapi-codegen` + `openapi-typescript`）は未着手です。

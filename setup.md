# セットアップ手順

## 構成

| レイヤ | 技術 | ポート |
| --- | --- | --- |
| リバースプロキシ | Caddy | 443 / 80 |
| フロントエンド | Vite + React + TypeScript + Tailwind CSS v4 + shadcn/ui | 3000 |
| バックエンド | Go (net/http) | 8080 |

Caddy が `https://localhost` で受けて、`/api/*` を Go(8080)、それ以外を Vite(3000) に転送します。

---

## 1. 必要なツール

```bash
# Caddy（軽量な方を入れる）
brew install caddy

# Go
brew install go

# Node.js（.nvmrc のバージョンを使う）
brew install nvm     # 未導入の場合
cd frontend && nvm install && nvm use

# pnpm（このプロジェクトのパッケージマネージャ）
corepack enable pnpm
```

`package.json` の `packageManager` で `pnpm@10.33.0` に固定しています。**npm / yarn は使わないでください**（lockfile が壊れます）。

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

## 3. フロントエンドの依存インストール

```bash
cd frontend
pnpm install
```

## 4. 起動

### 普段の開発（推奨）

**DBだけDocker、アプリはネイティブ**。Caddyは不要です。

```bash
# 1. DBを起動（バックグラウンド、1回でOK）
docker compose up -d db

# 2枚目のターミナル: バックエンド
cd backend && go run main.go

# 3枚目のターミナル: フロントエンド
cd frontend && pnpm dev
```

ブラウザで **http://localhost:3000** を開く。

`/api/*` は Vite の devProxy が Go(8080) に転送するので、Caddy を起動しなくても
API が叩けます。フロントだけ触る日は `pnpm dev` だけでも構いません
（API呼び出しは失敗しますが画面は出ます）。

DBを止めるときは `docker compose down`（データも消すなら `-v`）。

### 構成全体の確認（Caddy込み）

Caddyfile やプロキシ設定を変更した後の検証用です。

```bash
docker compose up -d      # 全サービス（frontend/backendはnginxダミー）
```

**http://localhost** で確認。終わったら `docker compose down`。

> ネイティブのCaddyとDockerのCaddyは同じ80番を使うため**同時に起動できません**。
> 切り替える際は `caddy stop` / `docker compose down` で必ず片方を止めてください。

### ネイティブのみ（ターミナル3枚）

```bash
# 1枚目: バックエンド
cd backend && go run main.go

# 2枚目: フロントエンド
cd frontend && pnpm dev

# 3枚目: リバースプロキシ（リポジトリルートで）
caddy run --config Caddyfile
```

ブラウザで **https://localhost** を開く。
「Call Go API」ボタンを押して `Hello, Go API!!` が表示されれば全レイヤの疎通OK。

> 初回は Caddy のローカル認証局を信頼させるため、管理者パスワードを求められます。

### 動作確認（CLI）

```bash
curl -k https://localhost/api/hello   # => {"message": "Hello, Go API!!"}
curl -k -o /dev/null -w "%{http_code}\n" https://localhost/   # => 200
```

---

## 5. フロントエンドの使い方

### コマンド

```bash
pnpm dev       # 開発サーバ (localhost:3000)
pnpm build     # 型チェック + 本番ビルド → dist/
pnpm lint      # oxlint
pnpm preview   # ビルド結果をローカル確認
```

### shadcn/ui のコンポーネント追加

必要になったものだけ都度追加してください。

```bash
cd frontend
pnpm dlx shadcn@latest add card input dialog
```

追加されたコンポーネントは `src/components/ui/` に**ソースとして**置かれるので、自由に編集できます。

### import エイリアス

`@/` が `src/` を指します。

```ts
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
```

### ディレクトリ

```
frontend/
├── .nvmrc                  # Node バージョン固定
├── components.json         # shadcn/ui 設定
├── vite.config.ts          # Tailwind プラグイン / @ エイリアス / port 3000
└── src/
    ├── index.css           # Tailwind + shadcn テーマ変数
    ├── App.tsx             # Go API 疎通サンプル
    ├── components/ui/      # shadcn/ui コンポーネント
    └── lib/utils.ts        # cn()
```

---

## トラブルシューティング

**`https://localhost` が意図しない内容を返す**
過去に `caddy start` したプロセスが残って 443 を掴んでいる可能性があります。

```bash
caddy stop
lsof -nP -iTCP:443 -sTCP:LISTEN    # 残っていれば kill <PID>
```

**ポート3000が埋まっている**
`vite.config.ts` の `server.port` と `Caddyfile` の転送先を揃えて変更してください。

---

## 6. Docker Compose（ヘルスチェック）

### 起動と状態確認

```bash
docker compose up -d
docker compose ps          # STATUS 列が (healthy) になればOK
```

`depends_on` に `condition: service_healthy` を付けているので、
**db → backend → frontend/backend → caddy** の順で「前段が healthy になってから」次が起動します。
「DBはまだ起動中なのにアプリが繋ぎにいって落ちる」という事故が防げます。

### ヘルスチェックの中身

| サービス | チェック方法 |
| --- | --- |
| db | `pg_isready` （Postgresが接続受付可能か） |
| backend | `wget --spider http://localhost/api/health` |
| frontend | `wget --spider http://localhost/` |
| caddy | `wget --spider http://localhost:2019/config/` （Caddyの管理API） |

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

### 疎通確認

```bash
curl http://localhost/              # => <h1>Frontend Ready!</h1>
curl http://localhost/api/health    # => {"status":"ok", ...}
```

### 停止

```bash
docker compose down          # コンテナ削除
docker compose down -v       # DBのデータも消す
```

---

## 7. Caddyfile について

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
| Docker（compose で上書き） | `:80` | `backend:80` | `frontend:80` |

**なぜ必要か**: コンテナ内の `localhost` はそのコンテナ自身を指すため、Docker では
`localhost:8080` ではなくサービス名 `backend` で名前解決する必要があります。
また `localhost { }` と書くとCaddyが自動HTTPS化して443にリダイレクトしますが、
composeでは80番しか公開していないため `:80` を指定しています。

---

## 補足 / TODO

- compose の frontend / backend はまだ **nginx のダミー**です。
  実際の Go / Vite に差し替える際は Dockerfile を用意してください。
  Go を `scratch` や `distroless` でビルドするとシェルが無く `CMD-SHELL` 形式の
  ヘルスチェックが使えないので、`alpine` ベースにするか、ヘルスチェック用の
  小さなGoバイナリを同梱する必要があります。

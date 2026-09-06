# hackson

ハッカソン用の開発環境です。

## 構成

| レイヤ | 技術 | ポート |
| --- | --- | --- |
| フロントエンド | Vite + React + TypeScript + Tailwind CSS + shadcn/ui | 3000 |
| バックエンド | Go (`net/http`) | 8080 |
| DB | PostgreSQL 16（Docker） | 5432 |
| リバースプロキシ | Caddy（構成確認用・普段は不要） | 80 / 443 |

## クイックスタート

事前に **Go / Docker Desktop / nvm** をインストールしてください。

```bash
# フロントエンドの準備（初回のみ）
cd frontend
nvm install && nvm use     # .nvmrc のバージョンに合わせる
corepack enable pnpm
pnpm install
cd ..

# 起動（ターミナル3枚）
docker compose up -d db        # 1. DB
cd backend && go run main.go   # 2. バックエンド
cd frontend && pnpm dev        # 3. フロントエンド
```

ブラウザで **http://localhost:3000** を開く。
「Call Go API」ボタンが動けば全レイヤ疎通OK。

停止するときは各ターミナルで `Ctrl+C`、DBは `docker compose down`。

## ドキュメント

| ファイル | 内容 |
| --- | --- |
| [setup.md](setup.md) | 環境構築の詳細・起動パターン・Docker/ヘルスチェック |
| [frontend/Frontend.md](frontend/Frontend.md) | フロントエンドの開発ルール |
| [backend/Backend.md](backend/Backend.md) | バックエンドの開発ルール |

## 注意

- パッケージマネージャは **pnpm** です。npm / yarn は使わないでください。
- API のパスは必ず `/api/` から始めてください（プロキシの振り分け条件）。

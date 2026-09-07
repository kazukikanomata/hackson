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

事前に **Docker Desktop** と **nvm** をインストールしてください。
バックエンドは Docker で動かすため、**Go のインストールは不要**です。

```bash
# 環境変数の準備（初回のみ）
cp .env.example .env

# フロントエンドの準備（初回のみ）
cd frontend
nvm install && nvm use     # .nvmrc のバージョンに合わせる
corepack enable pnpm
pnpm install
cd ..

# 起動（ターミナル2枚）
docker compose up -d db backend   # 1. DB + バックエンド（コード変更は自動反映）
cd frontend && pnpm dev           # 2. フロントエンド
```

ブラウザで **http://localhost:3000** を開く。
「Call Go API」ボタンで `Hello, Go API!!` が表示されれば疎通OKです。

> 初回だけバックエンドの起動に1〜2分かかります（開発ツールの取得のため）。
> 進捗は `docker compose logs -f backend` で確認できます。

停止は `docker compose down` と、フロントエンドのターミナルで `Ctrl+C`。

## ドキュメント

| ファイル | 内容 |
| --- | --- |
| [setup.md](setup.md) | 環境構築の詳細・起動パターン・ヘルスチェック・トラブルシューティング |
| [frontend/Frontend.md](frontend/Frontend.md) | フロントエンドの開発ルール |
| [backend/Backend.md](backend/Backend.md) | バックエンドの開発ルール |

## 開発上の約束

- パッケージマネージャは **pnpm** です。npm / yarn は使わないでください。
- APIのパスは必ず `/api/` から始めてください（プロキシの振り分け条件）。
- DBの認証情報は `.env` に置きます（git管理外）。**実際の値をコミットしないでください。**
  項目を増やしたら `.env.example` にも追記してください。
- `main` に直接 push しないでください。変更は feature branch から PR を出します
  （ドキュメントの修正のみ例外）。

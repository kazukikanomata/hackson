## 概要

フロントエンドの環境構築

| 項目                 | 採用技術                                   |
| -------------------- | ------------------------------------------ |
| ビルドツール         | Vite 8                                     |
| UI                   | React 19 + TypeScript 6                    |
| スタイル             | Tailwind CSS v4                            |
| UIコンポーネント     | shadcn/ui（Radixベース / nova プリセット） |
| パッケージマネージャ | pnpm 10                                    |
| 開発ポート           | 3000                                       |

`/api/*` へのリクエストは Caddy が Go(8080) に転送するため、フロントからは
`fetch('/api/hello')` のように相対パスで叩けます（CORS設定は不要）。

---

## セットアップ

```bash
# Node.js（.nvmrc のバージョンに合わせる）
nvm install && nvm use

# pnpm を有効化
corepack enable pnpm

# 依存をインストール
pnpm install
```

---

## コマンド

```bash
pnpm dev       # 開発サーバ起動（localhost:3000）
pnpm build     # 型チェック + 本番ビルド → dist/
pnpm lint      # oxlint
pnpm preview   # ビルド結果をローカルで確認
```

### API を叩く場合

`vite.config.ts` の devProxy が `/api/*` を Go(8080) に転送するので、
**Caddy は起動不要**です。Go だけ別ターミナルで動かしてください。

```bash
cd backend && go run main.go   # 別ターミナル
cd frontend && pnpm dev        # http://localhost:3000
```

フロントだけ触る日は `pnpm dev` のみでOKです（API呼び出しは失敗しますが画面は出ます）。
DBを使う場合はリポジトリルートで `docker compose up -d db`。

> `strictPort: true` を設定しているため、3000番が埋まっていると
> **別ポートにずれず起動失敗します**。先客を止めてから起動してください。
> ```bash
> lsof -ti TCP:3000 -sTCP:LISTEN | xargs kill
> ```

---

## ディレクトリ

```
frontend/
├── .nvmrc              # Node バージョン固定
├── components.json     # shadcn/ui の設定
├── vite.config.ts      # Tailwindプラグイン / @エイリアス / port 3000
└── src/
    ├── index.css       # Tailwind読み込み + shadcnのテーマ変数
    ├── App.tsx         # Go API 疎通サンプル
    ├── components/ui/  # shadcn/ui のコンポーネント
    └── lib/utils.ts    # cn()（クラス名結合ユーティリティ）
```

---

## 書き方のルール

### import エイリアス

`@/` が `src/` を指します。相対パスの `../../` は避けてください。

```ts
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
```

### UIコンポーネントの追加

必要になったものだけ、その都度追加します。

```bash
pnpm dlx shadcn@latest add card input dialog
```

追加分は `src/components/ui/` に**ソースコードとして**配置されるので、
そのまま編集して構いません（node_modules の中ではありません）。

利用できるコンポーネント一覧は https://ui.shadcn.com/docs/components を参照。

### スタイル

Tailwind のユーティリティクラスを使います。Tailwind v4 は設定ファイル
（`tailwind.config.js`）を持たず、`src/index.css` の CSS 変数でテーマを管理します。

色は `bg-background` / `text-foreground` / `text-muted-foreground` のような
shadcn のセマンティックな変数を使うと、ダークモード対応が自動で効きます。

---

## 注意点

- **ライブラリは必要になってから入れる。** 現状の依存は shadcn/ui の動作に必要な最小構成です。
  ルーターや状態管理はハッカソンの要綱が出てから判断します。
- `tsconfig` は `paths` のみでエイリアスを解決しています。TypeScript 6 で
  `baseUrl` が非推奨になったため、追加しないでください。

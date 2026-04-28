---
name: searxng-cli
description: "SearXNGの検索インスタンスに対してWeb検索を実行するCLIツール。コンパクトなJSON Lines形式で結果を返す。"
---

# SearXNG CLI

## 概要
SearXNG の検索インスタンスに対して Web 検索を実行する CLI ツール。
デフォルト出力は compact JSONL（1行1結果、キー短縮）。進捗は stderr に表示。

## 使い方
```bash
# 検索（デフォルト出力: compact JSONL）
searxng-cli search "<query>"

# カテゴリ指定（推奨）
searxng-cli search "<query>" -c general,news

# エンジン指定
searxng-cli search "<query>" --engines google,wikipedia

# 出力フォーマット変更
searxng-cli search "<query>" -f table
searxng-cli search "<query>" -f urls

# 件数制限（推奨）
searxng-cli search "<query>" --max-results 5

# 時間範囲
searxng-cli search "<query>" --time-range day

# タイムアウト変更（デフォルト30秒）
searxng-cli search "<query>" -t 60

# インスタンス切り替え
searxng-cli search "<query>" --instance myinst
```

## 出力形式（デフォルト: compact JSONL）
各行が 1 件の検索結果。キーは短縮（`t`=title, `u`=url, `s`=snippet, `c`=category, `e`=engine）。

```jsonl
{"t":"Title","u":"https://...","s":"snippet","c":"general","e":"google"}
```

## 動作
- stderr に `Searching <url> ... N results` と進捗を表示（stdout は結果のみ）
- 空結果の場合も `0 results` と表示
- 結果は stdout に出力（パイプ可能）

## 推奨フラグ
- `-c general,news` - カテゴリ指定（結果の品質向上）
- `--max-results 5` - トークン節約
- `--time-range day` - 最新情報取得時
- `-t 60` - SearXNGが遅い場合のタイムアウト延長

## 設定
初回実行前に設定が必要:
```bash
searxng-cli config init                       # デフォルト設定作成 (localhost:8080)
searxng-cli config set-instance public https://searx.example.com  # 公開インスタンス追加
searxng-cli config list                       # 設定一覧
```

環境変数 `SEARXNG_CLI_CONFIG_DIR` で設定ディレクトリを変更可能。

## 全フラグ一覧

| フラグ | エイリアス | デフォルト | 説明 |
|--------|-----------|-----------|------|
| `--format` | `-f` | `compact` | 出力形式: compact, table, urls, json |
| `--categories` | `-c` | `""` | カテゴリ指定 (例: general,news) |
| `--engines` | | `""` | エンジン指定 (例: google,wikipedia) |
| `--language` | | `""` | 言語コード (例: en, ja) |
| `--time-range` | | `""` | 期間: day, month, year |
| `--pageno` | `-n` | `1` | ページ番号 |
| `--instance` | | 設定のdefault | 使用するインスタンス名 |
| `--max-results` | | `0` (全て) | 表示件数上限 |
| `--timeout` | `-t` | `30` | リクエストタイムアウト(秒) |

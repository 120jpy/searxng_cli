# SearXNG CLI Skill

## 概要
SearXNG の検索インスタンスに対して Web 検索を実行する CLI ツール。
コンパクトな JSON Lines 形式で結果を返す。

## 使い方
```bash
# 検索（デフォルト出力: compact JSONL）
searxng-cli search "<query>"

# カテゴリ指定
searxng-cli search "<query>" -c general,news

# エンジン指定
searxng-cli search "<query>" --engines google,wikipedia

# 出力フォーマット変更
searxng-cli search "<query>" -f table
searxng-cli search "<query>" -f urls

# 件数制限
searxng-cli search "<query>" --max-results 5

# 時間範囲
searxng-cli search "<query>" --time-range day

# インスタンス切り替え
searxng-cli search "<query>" --instance myinst
```

## 出力形式（デフォルト: compact JSONL）

各行が 1 件の検索結果。キーは短縮（`t`=title, `u`=url, `s`=snippet, `c`=category, `e`=engine）。

```jsonl
{"t":"Title","u":"https://...","s":"snippet","c":"general","e":"google"}
```

## 推奨フラグ
- `-c general,news` - カテゴリ指定（結果の品質向上）
- `--max-results 5` - トークン節約
- `--time-range day` - 最新情報取得時

## 設定
初回実行前に設定が必要:
```bash
searxng-cli config init                       # デフォルト設定作成 (localhost:8888)
searxng-cli config set-instance public https://searx.example.com  # 公開インスタンス追加
searxng-cli config list                       # 設定一覧
```

環境変数 `SEARXNG_CLI_CONFIG_DIR` で設定ディレクトリを変更可能。

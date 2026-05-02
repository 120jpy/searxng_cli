# Release Procedure

```bash
# 1. バージョンタグを打つ
git tag v0.3.0

# 2. タグをプッシュ
git push origin master --tags

# 3. 全プラットフォームのバイナリをビルド
LDFLAGS="-X github.com/tomo/searxng-cli/cmd.Version=v0.3.0 \
         -X github.com/tomo/searxng-cli/cmd.Commit=$(git rev-parse --short HEAD) \
         -X github.com/tomo/searxng-cli/cmd.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

GOOS=darwin  GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/searxng-cli_darwin_amd64 .
GOOS=darwin  GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/searxng-cli_darwin_arm64 .
GOOS=linux   GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/searxng-cli_linux_amd64 .
GOOS=linux   GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/searxng-cli_linux_arm64 .
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/searxng-cli_windows_amd64.exe .

# 4. GitHub Release を作成（バイナリ添付）
gh release create v0.3.0 \
  --title "v0.3.0" \
  --notes "Release notes here" \
  dist/searxng-cli_darwin_amd64 \
  dist/searxng-cli_darwin_arm64 \
  dist/searxng-cli_linux_amd64 \
  dist/searxng-cli_linux_arm64 \
  dist/searxng-cli_windows_amd64.exe
```

## Makefile ターゲット（参考）

```bash
make build          # ビルド（バージョン情報付き）
make test           # テスト実行
make clean          # バイナリ削除
make release        # goreleaser 使用時（要 goreleaser + .goreleaser.yml）
make release-snapshot  # goreleaser 確認用
```

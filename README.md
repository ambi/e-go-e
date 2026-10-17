# e-go-e

Godog と Colly で、Web アプリケーションの E2E テストを書くサンプルリポジトリです。
Gherkin のキーワードは英語、テストの前提、操作、期待結果は日本語で記述し、`go test` から実行します。
サンプルアプリは Gin で作ったログイン画面です。
JavaScript やブラウザを起動せず、実際の HTTP 通信でフォーム、Cookie、CSRF、画面遷移を検証します。

## すぐに試す

Go 1.27.2 があれば、次のコマンドだけで実行できます。
Gin アプリがローカルの空きポートで自動起動し、テスト終了時に停止します。

```sh
go test -count=1 -v ./e2e
```

日本語のシナリオ本文と各ステップの実行結果を出力します。
Godog の CLI を別途インストールする必要はありません。
実行ログの保存には、Go 標準の出力をそのまま使えます。

```sh
mkdir -p artifacts
go test -count=1 -v ./e2e > artifacts/e2e.log 2>&1
```

## テスト内容を読む

開発者以外の方は、次の feature ファイルから確認内容を読めます。
この文章が Godog によって実行されるシナリオそのものです。

- [`e2e/features/login.feature`](e2e/features/login.feature)：ログイン成功、認証失敗、未ログイン時の5ケース。
- [`e2e/features/security.feature`](e2e/features/security.feature)：CSRF トークンがない場合と不正な場合の2ケース。

たとえばログイン成功は、次のように記述しています。

```gherkin
Scenario: 正しい認証情報でログインできる
  Given ログイン画面を開いている
  When 正しい認証情報でログインする
  Then マイページに利用者のメールアドレスが表示される
  And ページを開き直してもログイン状態が維持される
```

アプリを起動せず、ターミナルでシナリオだけを読むこともできます。

```sh
mise run e2e:spec
```

既存ステップの組み合わせで表せるケースは、feature ファイルにシナリオを追加します。
同じ操作で条件を変える場合は、CSRF の例のようにシナリオアウトラインと例の表を使えます。
新しい操作や検証が必要な場合は、[`e2e/steps_test.go`](e2e/steps_test.go) に処理を追加し、[`e2e/features_test.go`](e2e/features_test.go) でステップの文章と結び付けます。
未定義、未実装、曖昧なステップは `Strict: true` により失敗になります。

## 並行実行とシナリオの選択

各シナリオは新しい Colly Collector と CookieJar を使い、独立して最大4並行で実行します。
シナリオアウトラインの例の各行にも、個別の状態を用意します。
シナリオ内の操作は同期的に進め、検証に失敗した時点でそのシナリオを止めます。
`E2E_CONCURRENCY` で Godog の並行数を変更できます。
Go の `-parallel` フラグは Godog の並行数を変更しません。

```sh
E2E_CONCURRENCY=1 mise run e2e
E2E_CONCURRENCY=4 mise run e2e
```

Godog の並行実行ログはまとめて表示するため、Go のサブテストの進捗がシナリオ本文へ割り込むことを防いでいます。
特定のシナリオを Godog のタグで選択できます。

```sh
E2E_TAGS='@invalid-password' mise run e2e
E2E_TAGS='@smoke' mise run e2e
E2E_TAGS='@security' mise run e2e
```

`go test -v` ではシナリオ本文を表示する `pretty`、通常のテストでは簡潔な `progress` を使います。
`E2E_FORMAT=progress` などで変更できます。

## アプリとテストを別々に起動する

サンプルアプリを起動すると、`http://127.0.0.1:3000/login` でログイン画面を開けます。
ユーザーは `tester@example.com`、パスワードは `sample-password` です。
DB は使わず、サンプルユーザーとメモリ上のセッションを用意しています。

```sh
GIN_MODE=release go run ./cmd/sampleapp
```

別のターミナルで接続先とテスト用ユーザーを指定します。
明示した接続先にアクセスできなければ、テストは失敗します。

```sh
E2E_BASE_URL=http://127.0.0.1:3000 \
E2E_EMAIL=tester@example.com \
E2E_PASSWORD=sample-password \
go test -count=1 -v ./e2e
```

既存のアプリに移植する場合は、シナリオ内の URL、フォーム、入力欄の名前、期待する応答を合わせてください。
Colly はブラウザの入力検証を実行しないため、空のパスワードを送信するテストもサーバー側の検証です。

## Linux コンテナで実行する

Docker Engine または Docker Desktop と Compose v2 を使用します。
Gin アプリの起動完了をヘルスチェックで確認してから、別コンテナで E2E テストを実行します。

```sh
docker compose up --build --abort-on-container-exit --exit-code-from e2e
```

コマンドの終了コードはテスト結果に対応します。
実行後のコンテナの削除には、次のコマンドを使います。

```sh
docker compose down
```

テスト用イメージは Linux 向けのテストバイナリと CA 証明書を含みます。
既存のアプリへ接続する場合は、コンテナから到達できる URL を指定してください。

```sh
docker build --target e2e -t e-go-e .
docker run --rm \
  -e E2E_BASE_URL -e E2E_EMAIL -e E2E_PASSWORD \
  e-go-e
```

## 開発環境と lint

[mise](https://mise.jdx.dev/getting-started.html) 2026.10.12 以降を使う場合は、ツールの導入から確認まで次のコマンドで実行できます。
Go 1.27.2 と golangci-lint 2.15.0 は `mise.toml`、Go の依存は `go.mod` と `go.sum` で固定しています。

```sh
mise trust
mise install
mise run setup
mise run check
```

`golangci-lint` は `../idmagic` の設定を参考に、標準の検査、セキュリティ、エラー処理、context、テストヘルパーの検査を適用しています。
このサンプルで使わない DB 専用の検査と、参照元固有の除外設定は含めていません。
整形には同じく `gofumpt` を使い、追加の goimports は使っていません。

| コマンド | 内容 |
| --- | --- |
| `mise run check` | 整形確認、golangci-lint、race detector 付きテスト |
| `mise run lint-go` | golangci-lint の実行 |
| `mise run lint-go-fix` | 自動修正を含む lint |
| `mise run format-go` | gofumpt による整形 |
| `mise run e2e` | 日本語 Gherkin の本文と結果を表示して E2E を実行 |
| `mise run e2e:spec` | アプリを起動せず日本語シナリオを表示 |
| `mise run serve` | Gin サンプルアプリを起動 |
| `mise run build-linux` | 現在の CPU 向けに Linux 用バイナリをビルド |

## ファイル構成

```text
e2e/
  features/            # 非開発者も読める日本語 Gherkin
  features_test.go     # Godog の実行設定とステップ登録
  steps_test.go        # シナリオの各操作と検証
  helpers_test.go      # Colly の設定と画面確認
internal/sampleapp/
  app.go               # Gin の画面と認証処理
  templates/           # ログイン画面とマイページ
cmd/sampleapp/main.go   # サンプルアプリの起動
Dockerfile             # アプリ用とテスト用のイメージ
compose.yaml           # コンテナ間の E2E 実行
```

HTTP 通信、Cookie 管理、HTML の検索には Colly の機能を直接使っています。
フォームの hidden 項目を収集する処理は、サンプルのログイン画面用の短いヘルパーに留めています。
Colly の `AllowURLRevisit` と `ParseHTTPErrorResponse` を有効にし、同じ画面への再アクセスと 403／422 のエラー表示も検証します。
Collector の `Clone()` は CookieJar を共有するため、シナリオの分離には使いません。

feature ファイルはテストバイナリに埋め込むため、Linux のテスト用イメージでも同じシナリオを実行します。
Godog はシナリオを解釈して実行し、Colly が HTTP 操作を担当します。

- [Godog](https://github.com/cucumber/godog)
- [Colly の公式資料](https://pkg.go.dev/github.com/gocolly/colly/v2)
- [Gin](https://github.com/gin-gonic/gin)

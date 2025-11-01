# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## aws-tui

### 概要

k9sにインスパイアされたAWSサービス向けのターミナルUIアプリケーション。複数のAWSサービスをインタラクティブなTUIで操作可能。

**対応サービス:**
ACM, ACM PCA, CloudFront, CloudWatch, DynamoDB, EBS, EC2, ECS, EKS, ELB, ElastiCache, Global Accelerator, IAM, KMS, Lambda, MQ, MSK, RDS, Route 53, S3, SNS, SQS, Secrets Manager, Service Quotas, SSM, VPC

**主要技術:**
- Go 1.24.4
- AWS SDK for Go v2
- tview (TUIライブラリ)
- tcell (ターミナル制御)
- Cobra (CLIフレームワーク)

### 開発コマンド

```bash
# ビルド
go build -v ./...

# 実行
go run main.go

# 特定のパッケージのテスト
go test -v ./internal

# 全テスト実行
go test -v ./...

# カバレッジ付きテスト
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### AWS CLI依存

このアプリケーションは以下の機能でAWS CLIに依存しています:
- 現在のプロファイルのデフォルトリージョン取得
- プロファイル一覧取得（未実装）

### アーキテクチャ

#### パッケージ構造

```
aws-tui/
├── main.go                    # エントリポイント
├── cmd/
│   └── root.go               # Cobraルートコマンド定義
├── internal/
│   ├── app.go                # メインアプリケーションロジック
│   ├── header.go             # ヘッダーUI (アカウント情報など)
│   ├── footer.go             # フッターUI (キーバインド表示)
│   ├── key_action.go         # キーバインド定義
│   ├── <service>_<resource>.go  # 各AWSサービス/リソースのビューコンポーネント
│   ├── model/                # AWSリソースのモデル定義
│   ├── repo/                 # AWS SDK呼び出しのリポジトリ層
│   ├── ui/                   # 再利用可能なUIコンポーネント
│   │   ├── table.go          # テーブルUI
│   │   ├── text.go           # テキスト表示UI
│   │   └── tree.go           # ツリーUI
│   ├── view/                 # サービス別のビューインターフェース
│   ├── template/             # テンプレート処理
│   └── utils/                # ユーティリティ関数
```

#### レイヤー分離パターン

このアプリケーションは3層アーキテクチャを採用:

1. **View層** (`internal/<service>_<resource>.go`)
   - UIコンポーネントの実装
   - ユーザー入力の処理
   - キーバインドの定義
   - 例: `EC2Instances`, `RDSClusters`

2. **Repository層** (`internal/repo/`)
   - AWS SDKクライアントのラッパー
   - APIコール処理
   - ページネーション処理
   - 例: `repo.EC2`, `repo.RDS`

3. **Model層** (`internal/model/`)
   - AWS SDK型のエイリアス定義
   - データモデルのインターフェース
   - 例: `model.EC2Instance`, `model.RDSCluster`

#### Component インターフェース

すべてのビューは以下のインターフェースを実装:

```go
type Component interface {
    GetService() string        // サービス名を返す (例: "EC2")
    GetLabels() []string       // パンくずリスト用ラベル
    GetKeyActions() []KeyAction // キーバインド定義
    Render()                   // データをフェッチしてUIを更新
}
```

#### アプリケーションフロー

1. `main.go` → `cmd.Execute()` でCobraコマンド実行
2. `internal.NewApplication()` で:
   - AWS SDKクライアント初期化
   - Repositoryインスタンス作成
   - tviewアプリケーション構築 (Header/Pages/Footer)
3. ユーザー入力はグローバルキャプチャで処理:
   - `Esc`: ページを閉じる
   - `Ctrl+r`: リフレッシュ
   - その他: アクティブコンポーネントの `GetKeyActions()` に委譲
4. ビュー遷移は `app.AddAndSwitch(component)` で管理

#### UI/UXパターン

**ナビゲーション:**
- `g`: テーブルの先頭へ移動
- `G` (Shift+g): テーブルの末尾へ移動
- `j/k`: Vim風の上下移動 (一部ビューで)
- `Esc`: 前のビューに戻る

**ビュー遷移:**
ビューはスタック構造で管理され、`Application.pages` (tview.Pages) に追加/削除される。各ビュー名は `"{index} | {service} | {labels}"` 形式。

**データレンダリング:**
各コンポーネントの `Render()` メソッドで:
1. Repository経由でAWS APIからデータ取得
2. モデルをテーブル行データに変換
3. `ui.Table.SetData()` でテーブルを更新

#### リージョン管理

- デフォルト設定: `config.LoadDefaultConfig()` で自動検出
- Global Acceleratorなど一部サービスは `us-west-2` にハードコード
- 将来的にリージョン選択パネルの実装が予定されている (main.go のコメント参照)

### テスト

現在のテストカバレッジは限定的。テストファイルの例:
- `internal/key_action_test.go` - KeyActionの文字列表現テスト
- `internal/template/funcs_test.go` - テンプレート関数のテスト

テーブル駆動テストパターンを推奨。

### 新しいAWSサービスビューの追加方法

1. `internal/model/<service>.go` でモデル型を定義
2. `internal/repo/<service>.go` でリポジトリを実装
3. `internal/view/<service>.go` でサービスビューインターフェースを定義
4. `internal/<service>_<resource>.go` でコンポーネントを実装
5. `internal/app.go` の `NewApplication()` でクライアントとリポジトリを初期化
6. `repos` マップにリポジトリを追加

# grepo

[![Go Version](https://img.shields.io/badge/Go-1.25.3+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/ralsnet/grepo)](https://goreportcard.com/report/github.com/ralsnet/grepo)

**grepo**は、**ゼロ依存**で型安全なユースケース駆動型アプリケーションを構築するためのGoフレームワークです。Clean Architecture / Hexagonal Architectureパターンを実装し、ビジネスロジックと横断的関心事を明確に分離します。

## 🎯 特徴

### ✨ ゼロ依存
- **外部パッケージへの依存ゼロ**: コアライブラリは標準ライブラリのみを使用
- **軽量**: 最小限のフットプリントで高速な起動とビルド
- **セキュリティ**: 外部依存による脆弱性のリスクを排除
- **保守性**: 依存関係の更新やバージョン競合の心配不要

### 🔒 型安全
- **Goジェネリクス**: ユースケースの入出力に完全な型安全性を提供
- **コンパイル時チェック**: 実行前に型の不一致を検出

### ✅ 自動バリデーション
- **構造体タグベース**: `grepo:"optional"`, `grepo:"enum:admin,user"` などの宣言的なバリデーション
- **ゼロコンフィグ**: 必須フィールドチェックを自動実行
- **カスタマイズ可能**: 独自のバリデータを追加可能

### 🎣 フック機能
- **3階層のフック管理**: Root → Group → UseCaseの階層的な実行
- **BeforeHook**: 実行前処理（認証、ロギング、パラメータ変換）
- **AfterHook**: 実行後処理（メトリクス収集、監査ログ）
- **ErrorHook**: エラーハンドリング（アラート、エラーログ）

### 📋 API仕様生成
- **自動ドキュメント化**: 全ユースケースの入出力スキーマをJSON形式で出力
- **実行時イントロスペクション**: リフレクションによる型情報の抽出

### 🧪 テスト支援
- **時刻の注入**: `grepo.WithFixedTime()` で決定論的なテストを実現
- **モック可能**: インターフェースベースの設計で容易なモック化

## 📦 インストール

```bash
go get github.com/ralsnet/grepo
```

`go.mod` に追加:
```go
require github.com/ralsnet/grepo v0.1.0
```

## 🚀 クイックスタート

### 1. ユースケースを定義

```go
package usecase

import (
    "context"
    "github.com/ralsnet/grepo"
)

// 入力
type CreateUserInput struct {
    Name      string `grepo:""`                    // 必須フィールド
    Email     string `grepo:""`                    // 必須フィールド
    Role      string `grepo:"enum:admin,user"`     // enumバリデーション
    Age       *int   `grepo:"optional"`            // オプショナルフィールド
}

// 出力
type CreateUserOutput struct {
    UserID    string
    CreatedAt time.Time
}

// ユースケース実装
type CreateUser struct {
    userRepo UserRepository
}

func (uc *CreateUser) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
    // 現在時刻を取得（テスト時は固定値を注入可能）
    now := grepo.ExecuteTime(ctx)

    // ビジネスロジック
    userID := generateUserID()
    user := &User{
        ID:        userID,
        Name:      input.Name,
        Email:     input.Email,
        Role:      input.Role,
        CreatedAt: now,
    }

    if err := uc.userRepo.Save(ctx, user); err != nil {
        return nil, err
    }

    return &CreateUserOutput{
        UserID:    userID,
        CreatedAt: now,
    }, nil
}
```

### 2. APIを構築

```go
package main

import (
    "github.com/ralsnet/grepo"
    "github.com/ralsnet/grepo/hooks"
)

func main() {
    // グローバルフックを設定
    rootHook := grepo.NewGroupHook()
    rootHook.AddBefore(hooks.HookBeforeSlog())  // 全操作をログ
    rootHook.AddAfter(hooks.HookAfterSlog())
    rootHook.AddError(hooks.HookErrorSlog())

    // APIを構築
    api := grepo.NewAPIBuilder().
        WithHook(rootHook).
        WithOptions(
            grepo.WithEnableInputValidation(),   // 入力バリデーション有効化
            grepo.WithEnableOutputValidation(),  // 出力バリデーション有効化
        ).
        WithUseCase(grepo.NewUseCaseBuilder(&CreateUser{userRepo: repo}).Build()).
        Build()

    // ユースケースを実行
    ctx := context.Background()
    input := CreateUserInput{
        Name:  "山田太郎",
        Email: "taro@example.com",
        Role:  "admin",
    }

    output, err := grepo.UseCase[CreateUserInput, CreateUserOutput](api, "CreateUser").
        Execute(ctx, input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("ユーザー作成完了: ID=%s\n", output.UserID)
}
```

### 3. バリデーションエラーの処理

```go
// enumバリデーション違反
input := CreateUserInput{
    Name:  "山田太郎",
    Email: "taro@example.com",
    Role:  "superuser",  // ❌ "admin" または "user" のみ許可
}

_, err := grepo.UseCase[CreateUserInput, CreateUserOutput](api, "CreateUser").
    Execute(ctx, input)

// err は grepo.ErrValidation でラップされる
if errors.Is(err, grepo.ErrValidation) {
    fmt.Println("バリデーションエラー:", err)
}
```

## 🏗️ 主要コンポーネント

### API Registry ([api.go](api.go))
- 全ユースケースの中央レジストリ
- `Execute()` メソッドでフックライフサイクル全体を実行
- JSON形式でAPI仕様を出力可能

### UseCase実行エンジン ([usecase.go](usecase.go))
- `Executor[In, Out]` インターフェース: ビジネスロジックの抽象化
- `UseCaseBuilder`: フルエントAPIでユースケースを構築
- `WithHook()`, `WithGroup()` などで柔軟な設定

### バリデーション ([validate.go](validate.go))
- 構造体タグによる宣言的バリデーション
- `grepo:"optional"` - オプショナルフィールド
- `grepo:"enum:value1,value2"` - 列挙型制約
- カスタムバリデータの追加可能
- 再帰的に構造体と配列をバリデーション

### グループ管理 ([group.go](group.go))
- 名前付きフックのコレクション

### 標準フック ([hooks/hooks.go](hooks/hooks.go))
- `HookBeforeSlog()` - 操作開始のログ
- `HookAfterSlog()` - 成功完了のログ
- `HookErrorSlog()` - エラーログ
- カスタムフックの実装も可能

### コンテキストユーティリティ ([context.go](context.go))
- `ExecuteTime(ctx)` - 実行時刻を取得
- `WithFixedTime()` - テストに使用できる実行時刻の固定化

## 💡 ユースケース

### マイクロサービス
- Clean Architectureに基づいた構造化されたサービス設計
- ユースケースごとに明確な責任分離

### CQRS実装
- コマンドをユースケースとして実装
- 入出力の型安全性を保証

### CLIツール
- ビジネスロジックのバリデーション付きCLI
- `example/` ディレクトリに完全な実装例あり

### 監査ログ
- 全操作を自動的にログ記録
- フックを使った横断的なロギング

### API仕様生成
- ユースケースから自動的にJSON仕様を生成
- ドキュメント生成やフロントエンド連携に活用

## 📚 サンプル

`example/` ディレクトリに完全なユーザー管理システムの実装例があります:

```bash
cd example
go run ./cmd/cli spec              # API仕様を表示
go run ./cmd/cli get-user --id 123 # ユーザー取得
go run ./cmd/cli save-user --name "山田太郎" --authority admin
go run ./cmd/cli find-users        # 全ユーザー検索
```

## 📄 ライセンス

MIT License - 誰でも自由に使用できます。

詳細は [LICENSE](LICENSE) ファイルを参照してください。

---

**grepo** - ゼロ依存で型安全なClean Architectureフレームワーク

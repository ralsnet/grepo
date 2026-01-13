# grepo/cli

コマンドラインインターフェースを簡単に構築するためのCLIモジュールです。`grepo.API`から自動的にCLIコマンドを生成します。

## 特徴

- **自動コマンド生成**: APIのユースケースから自動的にCLIコマンドを生成
- **柔軟な入力方法**: JSON入力をファイル、標準入力、または引数から受け取り可能
- **型安全**: reflectionを使用して構造体型を保持したままJSON入力を処理
- **スキーマ表示**: 各コマンドのInput/Outputスキーマをヘルプで確認可能
- **API仕様の出力**: `spec`コマンドで全API仕様をJSON形式で出力

## インストール

```bash
go get github.com/ralsnet/grepo/cli
```

## 使い方

### 基本的な使用例

```go
package main

import (
    "fmt"
    "os"

    "github.com/ralsnet/grepo/cli"
    "github.com/ralsnet/grepo/example/internal"
)

func main() {
    api := internal.InitializeAPI()
    if err := cli.New(api, "myapp").Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}
```

### コマンドの実行

#### ヘルプの表示

```bash
# 全コマンドのリスト
$ myapp --help

# 特定コマンドのヘルプとスキーマ
$ myapp GetUser --help
```

#### JSON入力の方法

**1. 標準入力から**

```bash
$ echo '{"ID":"user1"}' | myapp GetUser
```

**2. ファイルから**

```bash
$ myapp GetUser -i input.json
# または
$ myapp GetUser --input input.json
```

**3. コマンドライン引数から**

```bash
$ myapp GetUser '{"ID":"user1"}'
```

#### API仕様の確認

```bash
# 全API仕様をJSON形式で出力
$ myapp spec
```

出力例:
```json
{
  "GetUser": {
    "Operation": "GetUser",
    "Input": {
      "Kind": "object",
      "Name": "usecase.GetUserInput",
      "Fields": [
        {
          "Field": "ID",
          "Type": {
            "Kind": "string",
            "Name": "string"
          }
        }
      ]
    },
    "Output": { ... }
  }
}
```

## 実装の詳細

### 入力の型変換

`Descriptor.Input()`から返される`any`型を、reflectionを使って適切な構造体型に変換します：

```go
p := reflect.New(reflect.TypeOf(uc.Input())).Interface()
if err := json.Unmarshal(b, p); err != nil {
    return nil, err
}
return reflect.ValueOf(p).Elem().Interface(), nil
```

この方法により、`json.Unmarshal`でmap型になることを回避し、元の構造体型を保持します。

### 自動生成されるコマンド

- 各UseCaseは自動的にコマンドとして登録されます
- コマンド名は`Descriptor.Operation()`から取得
- 説明文は`Descriptor.Description()`から取得
- Input/Outputスキーマは`--help`で自動表示

## 完全な例

[example/cmd/clitest](../example/cmd/clitest)ディレクトリに完全な動作例があります。

```bash
# 実行例
cd example/cmd/clitest

# ユーザー作成
echo '{"Name":"TestUser","Authority":"admin"}' | go run main.go SaveUser

# ユーザー取得
echo '{"ID":"019bb5be-cfa3-7f05-8434-f5e3a55dc73b"}' | go run main.go GetUser

# ユーザー検索
echo '{"Name":"Test"}' | go run main.go FindUsers

# API仕様の確認
go run main.go spec
```

## ライセンス

このモジュールは[grepo](https://github.com/ralsnet/grepo)プロジェクトの一部です。

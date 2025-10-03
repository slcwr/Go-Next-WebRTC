# ユニットテスト ガイド

Go標準の`testing`パッケージを使用したユニットテストの実装。

## 📁 ファイル構成

```
backend/
├── internal/application/usecase/
│   ├── auth_usecase.go
│   ├── auth_usecase_test.go          # 認証ユースケースのテスト
│   ├── todo_usecase.go
│   ├── todo_usecase_test.go          # Todoユースケースのテスト
│   └── testutil/
│       ├── mock_repositories.go       # モックリポジトリ実装
│       └── mock_todo_repository.go
└── pkg/jwt/
    ├── jwt.go
    └── jwt_test.go                    # JWTサービスのテスト
```

## 🧪 テストの実行

### 全テスト実行
```bash
cd backend
go test ./...
```

### 特定パッケージのテスト
```bash
# UseCaseのテスト
go test ./internal/application/usecase/...

# JWTのテスト
go test ./pkg/jwt/...
```

### 詳細出力（-v）
```bash
go test -v ./internal/application/usecase/...
```

### カバレッジ測定
```bash
# カバレッジレポート生成
go test -cover ./...

# 詳細なカバレッジレポート
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 特定のテスト実行
```bash
# テスト名でフィルタ
go test -v ./internal/application/usecase/... -run TestAuthUseCase_Register

# サブテストの実行
go test -v ./internal/application/usecase/... -run TestAuthUseCase_Register/valid_registration
```

## 📝 テストの構造

### Table-Driven Tests（テーブル駆動テスト）

```go
func TestAuthUseCase_Register(t *testing.T) {
    tests := []struct {
        name        string
        email       string
        password    string
        userName    string
        wantErr     bool
        expectedErr error
    }{
        {
            name:     "valid registration",
            email:    "test@example.com",
            password: "ValidPass123!",
            userName: "Test User",
            wantErr:  false,
        },
        {
            name:        "empty email",
            email:       "",
            password:    "ValidPass123!",
            userName:    "Test User",
            wantErr:     true,
            expectedErr: entity.ErrEmailRequired,
        },
        // 他のテストケース...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実装
        })
    }
}
```

### AAA パターン（Arrange-Act-Assert）

```go
func TestTodoUsecase_CreateTodo(t *testing.T) {
    // Arrange（準備）
    repo := testutil.NewMockTodoRepository()
    usecase := NewTodoUsecase(repo)
    ctx := context.Background()

    // Act（実行）
    todo, err := usecase.CreateTodo(ctx, 1, "Test Todo", "Description")

    // Assert（検証）
    if err != nil {
        t.Errorf("CreateTodo() unexpected error = %v", err)
    }
    if todo.Title != "Test Todo" {
        t.Errorf("CreateTodo() Title = %v, want Test Todo", todo.Title)
    }
}
```

## 🎭 モックの使用

### モックリポジトリの作成

```go
// デフォルトの振る舞い
repo := testutil.NewMockUserRepository()

// カスタムの振る舞いを設定
repo.FindByEmailFunc = func(ctx context.Context, email string) (*entity.User, error) {
    return nil, errors.New("database error")
}
```

### モックの利点
- データベース不要で高速
- エッジケースのテストが容易
- 依存関係の分離

## 📊 テストカバレッジ

### 既存のテストケース

#### AuthUseCase（auth_usecase_test.go）
- ✅ Register - 新規登録（有効/無効な入力）
- ✅ Register - メール重複チェック
- ✅ Login - ログイン（正常/異常系）
- ✅ RefreshToken - トークン更新
- ✅ Logout - ログアウト
- ✅ GetUserByID - ユーザー取得
- ✅ ChangePassword - パスワード変更

#### TodoUsecase（todo_usecase_test.go）
- ✅ CreateTodo - Todo作成
- ✅ GetAllTodos - 全Todo取得（ユーザー分離）
- ✅ GetTodoByID - 個別Todo取得
- ✅ UpdateTodo - Todo更新（タイトル/説明/完了状態）
- ✅ DeleteTodo - Todo削除

#### JWT Service（jwt_test.go）
- ✅ GenerateAccessToken - トークン生成
- ✅ ValidateAccessToken - トークン検証
- ✅ トークン有効期限テスト
- ✅ 異なるシークレットでの検証失敗
- ✅ クレーム内容の検証

## 🔍 テストケースの例

### 正常系テスト
```go
func TestAuthUseCase_Register_Valid(t *testing.T) {
    repo := testutil.NewMockUserRepository()
    authRepo := testutil.NewMockAuthRepository()
    config := NewAuthConfig("test-secret-key")
    usecase := NewAuthUseCase(repo, authRepo, config)

    tokens, err := usecase.Register(
        context.Background(),
        "test@example.com",
        "ValidPass123!",
        "Test User",
    )

    if err != nil {
        t.Fatalf("Register() failed: %v", err)
    }
    if tokens.AccessToken == "" {
        t.Error("AccessToken is empty")
    }
}
```

### 異常系テスト
```go
func TestAuthUseCase_Register_InvalidEmail(t *testing.T) {
    usecase := setupAuthUseCase()

    _, err := usecase.Register(
        context.Background(),
        "invalid-email",
        "ValidPass123!",
        "Test User",
    )

    if err == nil {
        t.Error("Register() should fail with invalid email")
    }
}
```

### エッジケーステスト
```go
func TestTodoUsecase_UpdateTodo_OtherUser(t *testing.T) {
    repo := testutil.NewMockTodoRepository()
    usecase := NewTodoUsecase(repo)

    // ユーザー1のTodoを作成
    todo, _ := usecase.CreateTodo(context.Background(), 1, "Todo", "Desc")

    // ユーザー2が更新を試みる
    newTitle := "Hacked"
    _, err := usecase.UpdateTodo(context.Background(), todo.ID, 2, &newTitle, nil, nil)

    if err == nil {
        t.Error("Should not allow updating other user's todo")
    }
}
```

## 🛠️ テストのベストプラクティス

### 1. テスト名は説明的に
```go
// ❌ Bad
func TestRegister(t *testing.T) {}

// ✅ Good
func TestAuthUseCase_Register_ValidInput(t *testing.T) {}
func TestAuthUseCase_Register_DuplicateEmail(t *testing.T) {}
```

### 2. 各テストは独立させる
```go
// ❌ Bad - テストが他のテストに依存
var globalUser *entity.User
func TestA(t *testing.T) {
    globalUser = createUser()
}
func TestB(t *testing.T) {
    // globalUserに依存
}

// ✅ Good - 各テストで準備
func TestA(t *testing.T) {
    user := createUser()
    // ...
}
```

### 3. エラーメッセージは具体的に
```go
// ❌ Bad
if err != nil {
    t.Error("failed")
}

// ✅ Good
if err != nil {
    t.Errorf("CreateTodo() unexpected error = %v, input: %+v", err, input)
}
```

### 4. サブテストを活用
```go
t.Run("with valid input", func(t *testing.T) {
    // テスト
})
t.Run("with invalid input", func(t *testing.T) {
    // テスト
})
```

## 🚀 継続的インテグレーション

### GitHub Actionsでの自動テスト
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html
```

## 📚 参考資料

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table-Driven Tests in Go](https://go.dev/wiki/TableDrivenTests)
- [Test Coverage](https://go.dev/blog/cover)

## ✅ テストチェックリスト

実装時に確認すべき項目：

- [ ] 正常系のテストがある
- [ ] 異常系のテストがある
- [ ] エッジケースのテストがある
- [ ] エラーメッセージが具体的
- [ ] テスト名が説明的
- [ ] 各テストが独立している
- [ ] モックを適切に使用している
- [ ] カバレッジが適切（目標: 80%以上）

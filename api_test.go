package grepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

// テスト用の入力・出力型
type TestInput struct {
	Value int `json:"value"`
}

type TestOutput struct {
	Result int `json:"result"`
}

// テスト用のユースケース実装
type addOneUseCase struct{}

func (u *addOneUseCase) Execute(ctx context.Context, input TestInput) (*TestOutput, error) {
	return &TestOutput{Result: input.Value + 1}, nil
}

type errorUseCase struct{}

func (u *errorUseCase) Execute(ctx context.Context, input TestInput) (*TestOutput, error) {
	return nil, errors.New("test error")
}

type validatedInput struct {
	Value int `json:"value"`
}

type validatedUseCase struct{}

func (u *validatedUseCase) Execute(ctx context.Context, input validatedInput) (*TestOutput, error) {
	return &TestOutput{Result: input.Value * 2}, nil
}

func TestAPI_ExecuteAny(t *testing.T) {
	tests := []struct {
		name      string
		setupAPI  func() *API
		operation string
		input     any
		want      any
		wantErr   bool
		errType   error
	}{
		{
			name: "正常系: ユースケース実行成功",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					Build()
			},
			operation: "add_one",
			input:     TestInput{Value: 5},
			want:      &TestOutput{Result: 6},
			wantErr:   false,
		},
		{
			name: "異常系: 存在しないオペレーション",
			setupAPI: func() *API {
				return NewAPIBuilder().Build()
			},
			operation: "not_exists",
			input:     TestInput{Value: 5},
			want:      nil,
			wantErr:   true,
			errType:   ErrNotFound,
		},
		{
			name: "異常系: ユースケースがエラーを返す",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&errorUseCase{}).
					WithOperation("error_uc").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					Build()
			},
			operation: "error_uc",
			input:     TestInput{Value: 5},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "正常系: 固定時刻オプション",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				return NewAPIBuilder().
					AddUseCase(uc).
					WithOptions(WithFixedTime(fixedTime)).
					Build()
			},
			operation: "add_one",
			input:     TestInput{Value: 10},
			want:      &TestOutput{Result: 11},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := tt.setupAPI()
			got, err := api.ExecuteAny(context.Background(), tt.operation, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteAny() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("ExecuteAny() error = %v, wantErrType %v", err, tt.errType)
				return
			}

			if !tt.wantErr {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("ExecuteAny() = %v, want %v", string(gotJSON), string(wantJSON))
				}
			}
		})
	}
}

func TestAPI_UseCases(t *testing.T) {
	tests := []struct {
		name      string
		setupAPI  func() *API
		wantCount int
		wantOps   []string
	}{
		{
			name: "正常系: 複数のユースケースが登録されている",
			setupAPI: func() *API {
				uc1 := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				uc2 := NewUseCaseBuilder(&errorUseCase{}).
					WithOperation("error_uc").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc1).
					AddUseCase(uc2).
					Build()
			},
			wantCount: 2,
			wantOps:   []string{"add_one", "error_uc"},
		},
		{
			name: "正常系: ユースケースが登録されていない",
			setupAPI: func() *API {
				return NewAPIBuilder().Build()
			},
			wantCount: 0,
			wantOps:   []string{},
		},
		{
			name: "正常系: ソート順序の確認",
			setupAPI: func() *API {
				uc1 := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("z_last").
					Build()
				uc2 := NewUseCaseBuilder(&errorUseCase{}).
					WithOperation("a_first").
					Build()
				uc3 := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("m_middle").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc1).
					AddUseCase(uc2).
					AddUseCase(uc3).
					Build()
			},
			wantCount: 3,
			wantOps:   []string{"a_first", "m_middle", "z_last"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := tt.setupAPI()
			got := api.UseCases()

			if len(got) != tt.wantCount {
				t.Errorf("UseCases() count = %v, want %v", len(got), tt.wantCount)
				return
			}

			for i, op := range tt.wantOps {
				if got[i].Operation() != op {
					t.Errorf("UseCases()[%d].Operation() = %v, want %v", i, got[i].Operation(), op)
				}
			}
		})
	}
}

func TestAPI_WithValidation(t *testing.T) {
	tests := []struct {
		name      string
		setupAPI  func() *API
		operation string
		input     any
		wantErr   bool
	}{
		{
			name: "正常系: 入力バリデーションが有効",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&validatedUseCase{}).
					WithOperation("validated").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					WithOptions(WithEnableInputValidation()).
					Build()
			},
			operation: "validated",
			input:     validatedInput{Value: 50},
			wantErr:   false,
		},
		{
			name: "正常系: 出力バリデーションが有効",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&validatedUseCase{}).
					WithOperation("validated").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					WithOptions(WithEnableOutputValidation()).
					Build()
			},
			operation: "validated",
			input:     validatedInput{Value: 50},
			wantErr:   false,
		},
		{
			name: "正常系: バリデーション無効",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&validatedUseCase{}).
					WithOperation("validated").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					Build()
			},
			operation: "validated",
			input:     validatedInput{Value: 0},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := tt.setupAPI()
			_, err := api.ExecuteAny(context.Background(), tt.operation, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteAny() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPI_WithHooks(t *testing.T) {
	tests := []struct {
		name      string
		setupAPI  func(*testing.T) (*API, *[]string)
		operation string
		input     any
		want      *TestOutput
		wantOrder []string
		wantErr   bool
	}{
		{
			name: "正常系: Beforeフック実行",
			setupAPI: func(t *testing.T) (*API, *[]string) {
				order := &[]string{}
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				api := NewAPIBuilder().
					AddUseCase(uc).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before")
						return ctx, nil
					}).
					Build()
				return api, order
			},
			operation: "add_one",
			input:     TestInput{Value: 5},
			want:      &TestOutput{Result: 6},
			wantOrder: []string{"before"},
			wantErr:   false,
		},
		{
			name: "正常系: Afterフック実行",
			setupAPI: func(t *testing.T) (*API, *[]string) {
				order := &[]string{}
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				api := NewAPIBuilder().
					AddUseCase(uc).
					AddAfterHook(func(ctx context.Context, desc Descriptor, i any, o any) {
						*order = append(*order, "after")
					}).
					Build()
				return api, order
			},
			operation: "add_one",
			input:     TestInput{Value: 5},
			want:      &TestOutput{Result: 6},
			wantOrder: []string{"after"},
			wantErr:   false,
		},
		{
			name: "正常系: Errorフック実行",
			setupAPI: func(t *testing.T) (*API, *[]string) {
				order := &[]string{}
				uc := NewUseCaseBuilder(&errorUseCase{}).
					WithOperation("error_uc").
					Build()
				api := NewAPIBuilder().
					AddUseCase(uc).
					AddErrorHook(func(ctx context.Context, desc Descriptor, i any, e error) {
						*order = append(*order, "error")
					}).
					Build()
				return api, order
			},
			operation: "error_uc",
			input:     TestInput{Value: 5},
			want:      nil,
			wantOrder: []string{"error"},
			wantErr:   true,
		},
		{
			name: "正常系: 複数のフックが順番に実行される",
			setupAPI: func(t *testing.T) (*API, *[]string) {
				order := &[]string{}
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				api := NewAPIBuilder().
					AddUseCase(uc).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before1")
						return ctx, nil
					}).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before2")
						return ctx, nil
					}).
					AddAfterHook(func(ctx context.Context, desc Descriptor, i any, o any) {
						*order = append(*order, "after1")
					}).
					AddAfterHook(func(ctx context.Context, desc Descriptor, i any, o any) {
						*order = append(*order, "after2")
					}).
					Build()
				return api, order
			},
			operation: "add_one",
			input:     TestInput{Value: 5},
			want:      &TestOutput{Result: 6},
			wantOrder: []string{"before1", "before2", "after1", "after2"},
			wantErr:   false,
		},
		{
			name: "正常系: ユースケースフック",
			setupAPI: func(t *testing.T) (*API, *[]string) {
				order := &[]string{}
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					AddBeforeHook(func(ctx context.Context, i *TestInput) (context.Context, error) {
						i.Value = 1000
						*order = append(*order, "before3")
						return ctx, nil
					}).
					Build()
				api := NewAPIBuilder().
					AddUseCase(uc).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before1")
						return ctx, nil
					}).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before2")
						return ctx, nil
					}).
					AddAfterHook(func(ctx context.Context, desc Descriptor, i any, o any) {
						*order = append(*order, "after1")
					}).
					AddAfterHook(func(ctx context.Context, desc Descriptor, i any, o any) {
						*order = append(*order, "after2")
					}).
					Build()
				return api, order
			},
			operation: "add_one",
			input:     TestInput{Value: 5},
			want:      &TestOutput{Result: 1001},
			wantOrder: []string{"before1", "before2", "before3", "after1", "after2"},
			wantErr:   false,
		},
		{
			name: "異常系: ユースケースフックでエラー",
			setupAPI: func(t *testing.T) (*API, *[]string) {
				order := &[]string{}
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					AddBeforeHook(func(ctx context.Context, i *TestInput) (context.Context, error) {
						*order = append(*order, "before3")
						return ctx, fmt.Errorf("error in before hook")
					}).
					AddErrorHook(func(ctx context.Context, i TestInput, err error) {
						*order = append(*order, "error2")
					}).
					Build()
				api := NewAPIBuilder().
					AddUseCase(uc).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before1")
						return ctx, nil
					}).
					AddBeforeHook(func(ctx context.Context, desc Descriptor, i any) (context.Context, error) {
						*order = append(*order, "before2")
						return ctx, nil
					}).
					AddErrorHook(func(ctx context.Context, desc Descriptor, i any, e error) {
						*order = append(*order, "error1")
					}).
					Build()
				return api, order
			},
			operation: "add_one",
			input:     TestInput{Value: 5},
			want:      nil,
			wantOrder: []string{"before1", "before2", "before3", "error1", "error2"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, order := tt.setupAPI(t)
			output, err := api.ExecuteAny(context.Background(), tt.operation, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteAny() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && tt.want.Result != output.(*TestOutput).Result {
				t.Errorf("ExecuteAny() = %v, want %v", output, tt.want)
			}

			if len(*order) != len(tt.wantOrder) {
				t.Errorf("Hook execution order length = %v, want %v", len(*order), len(tt.wantOrder))
				return
			}

			for i, want := range tt.wantOrder {
				if (*order)[i] != want {
					t.Errorf("Hook execution order[%d] = %v, want %v", i, (*order)[i], want)
				}
			}
		})
	}
}

func TestAPIBuilder(t *testing.T) {
	tests := []struct {
		name     string
		buildAPI func() *API
		validate func(*testing.T, *API)
	}{
		{
			name: "正常系: 説明付きAPIの構築",
			buildAPI: func() *API {
				return NewAPIBuilder().
					WithDescription("Test API").
					Build()
			},
			validate: func(t *testing.T, api *API) {
				if api.Description() != "Test API" {
					t.Errorf("Description() = %v, want %v", api.Description(), "Test API")
				}
			},
		},
		{
			name: "正常系: 複数オプション付きAPIの構築",
			buildAPI: func() *API {
				fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				return NewAPIBuilder().
					WithDescription("Test API with Options").
					WithOptions(
						WithFixedTime(fixedTime),
						WithEnableInputValidation(),
						WithEnableOutputValidation(),
					).
					Build()
			},
			validate: func(t *testing.T, api *API) {
				if api.Description() != "Test API with Options" {
					t.Errorf("Description() = %v, want %v", api.Description(), "Test API with Options")
				}
				if !api.options.enableInputValidation {
					t.Error("Input validation should be enabled")
				}
				if !api.options.enableOutputValidation {
					t.Error("Output validation should be enabled")
				}
				if api.options.fixedTime == nil {
					t.Error("Fixed time should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := tt.buildAPI()
			tt.validate(t, api)
		})
	}
}

func TestUseCase(t *testing.T) {
	tests := []struct {
		name      string
		setupAPI  func() *API
		operation string
		input     TestInput
		want      *TestOutput
		wantErr   bool
	}{
		{
			name: "正常系: UseCaseヘルパー関数",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					Build()
			},
			operation: "add_one",
			input:     TestInput{Value: 10},
			want:      &TestOutput{Result: 11},
			wantErr:   false,
		},
		{
			name: "異常系: 存在しないオペレーション",
			setupAPI: func() *API {
				return NewAPIBuilder().Build()
			},
			operation: "not_exists",
			input:     TestInput{Value: 10},
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := tt.setupAPI()
			executor := UseCase[TestInput, TestOutput](api, tt.operation)
			got, err := executor.Execute(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("UseCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("UseCase() = %v, want %v", string(gotJSON), string(wantJSON))
				}
			}
		})
	}
}

func TestUseCaseByIO(t *testing.T) {
	tests := []struct {
		name     string
		setupAPI func() *API
		input    TestInput
		want     *TestOutput
		wantErr  bool
	}{
		{
			name: "正常系: 型による検索",
			setupAPI: func() *API {
				uc := NewUseCaseBuilder(&addOneUseCase{}).
					WithOperation("add_one").
					Build()
				return NewAPIBuilder().
					AddUseCase(uc).
					Build()
			},
			input:   TestInput{Value: 20},
			want:    &TestOutput{Result: 21},
			wantErr: false,
		},
		{
			name: "異常系: 該当するユースケースがない",
			setupAPI: func() *API {
				return NewAPIBuilder().Build()
			},
			input:   TestInput{Value: 20},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := tt.setupAPI()
			executor := UseCaseByIO[TestInput, TestOutput](api)
			got, err := executor.Execute(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("UseCaseByIO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("UseCaseByIO() = %v, want %v", string(gotJSON), string(wantJSON))
				}
			}
		})
	}
}

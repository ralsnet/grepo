package grepo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

type APIOptions struct {
	fixedTime              *time.Time
	enableInputValidation  bool
	enableOutputValidation bool
	customFieldValidators  []FieldValidator
}

type APIOptionFunc func(*APIOptions)

func WithFixedTime(t time.Time) APIOptionFunc {
	return func(o *APIOptions) {
		o.fixedTime = &t
	}
}

func WithEnableInputValidation() APIOptionFunc {
	return func(o *APIOptions) {
		o.enableInputValidation = true
	}
}

func WithEnableOutputValidation() APIOptionFunc {
	return func(o *APIOptions) {
		o.enableOutputValidation = true
	}
}

func WithCustomFieldValidators(validators ...FieldValidator) APIOptionFunc {
	return func(o *APIOptions) {
		o.customFieldValidators = append(o.customFieldValidators, validators...)
	}
}

type API struct {
	description string
	m           map[string]Descriptor
	root        *Group
	options     *APIOptions
}

func newAPI() *API {
	return &API{
		m:       make(map[string]Descriptor),
		root:    NewGroup("root"),
		options: &APIOptions{},
	}
}

func UseCase[I any, O any](api *API, op string) Executor[I, O] {
	return (ExecutorFunc[I, O](func(ctx context.Context, input I) (*O, error) {
		out, err := api.ExecuteAny(ctx, op, input)
		if err != nil {
			return nil, err
		}
		output, ok := out.(*O)
		if !ok {
			return nil, fmt.Errorf("invalid output type")
		}
		return output, nil
	}))
}

func UseCaseByIO[I any, O any](api *API) Executor[I, O] {
	return (ExecutorFunc[I, O](func(ctx context.Context, input I) (*O, error) {
		var uc Descriptor
		for _, d := range api.UseCases() {
			interactor, ok := d.(*Interactor[I, O])
			if ok {
				uc = interactor
				break
			}
		}
		if uc == nil {
			return nil, ErrNotFound
		}
		out, err := api.executeUseCase(ctx, uc, input)
		if err != nil {
			return nil, err
		}
		output, ok := out.(*O)
		if !ok {
			return nil, fmt.Errorf("invalid output type")
		}
		return output, nil
	}))
}

func (a *API) Description() string {
	return a.description
}

func (a *API) UseCases() []Descriptor {
	keys := make([]string, 0, len(a.m))
	for key := range a.m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	descs := make([]Descriptor, 0, len(a.m))
	for _, k := range keys {
		descs = append(descs, a.m[k])
	}
	return descs
}

func (a *API) ExecuteAny(ctx context.Context, operation string, input any) (any, error) {
	uc, ok := a.m[operation]
	if !ok {
		return nil, ErrNotFound
	}
	output, err := a.executeUseCase(ctx, uc, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (a *API) executeUseCase(ctx context.Context, uc Descriptor, input any) (output any, err error) {
	interactorType := reflect.TypeOf(uc)
	interactorValue := reflect.ValueOf(uc)
	for interactorType.Kind() == reflect.Pointer {
		interactorType = interactorType.Elem()
	}

	// Create a pointer to input value
	ptr := reflect.New(reflect.ValueOf(input).Type())
	ptr.Elem().Set(reflect.ValueOf(input))
	inputPtr := ptr.Interface()

	groups := append([]*Group{a.root}, uc.Groups()...)

	defer func() {
		if err != nil {
			output = nil
			hookError(ctx, uc, ptr.Elem().Interface(), err, groups)
			a.doErrorHook(ctx, interactorValue, ptr.Elem().Interface(), err)
		}
	}()

	if a.options.fixedTime != nil {
		ctx = WithExecuteTime(ctx, *a.options.fixedTime)
	} else {
		ctx = WithExecuteTime(ctx, time.Now())
	}

	ctx, err = hookBefore(ctx, uc, ptr.Elem().Interface(), groups)
	if err != nil {
		return nil, err
	}

	ctx, err = a.doBeforeHook(ctx, interactorValue, inputPtr)
	if err != nil {
		return nil, err
	}

	input = ptr.Elem().Interface()

	if a.options.enableInputValidation {
		if err = Validate(input, a.options.customFieldValidators...); err != nil {
			return nil, err
		}
	}

	execute := interactorValue.MethodByName("Execute")
	if !execute.IsValid() {
		return nil, ErrNotFound
	}
	o := execute.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(input)})
	if len(o) != 2 {
		return nil, ErrInvalid
	}

	output = o[0].Interface()
	err, _ = o[1].Interface().(error)
	if err != nil {
		return nil, err
	}

	if a.options.enableOutputValidation {
		if err = Validate(output, a.options.customFieldValidators...); err != nil {
			return nil, err
		}
	}

	hookAfter(ctx, uc, input, output, groups)
	a.doAfterHook(ctx, interactorValue, input, output)
	return output, nil
}

func (a *API) doBeforeHook(ctx context.Context, interactor reflect.Value, input any) (context.Context, error) {
	doBeforeHook := interactor.MethodByName("DoBeforeHook")
	if !doBeforeHook.IsValid() {
		return nil, ErrNotFound
	}

	o := doBeforeHook.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(input)})
	if len(o) != 2 {
		return nil, ErrInvalid
	}
	ctxInterface, ok := o[0].Interface().(context.Context)
	if !ok {
		return ctx, ErrInvalid
	}
	ctx = ctxInterface
	errInterface, ok := o[1].Interface().(error)
	if ok {
		return ctx, errInterface
	}
	return ctx, nil
}

func (a *API) doAfterHook(ctx context.Context, interactor reflect.Value, input any, output any) {
	doAfterHook := interactor.MethodByName("DoAfterHook")
	if !doAfterHook.IsValid() {
		return
	}
	doAfterHook.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(input), reflect.ValueOf(output)})
}

func (a *API) doErrorHook(ctx context.Context, interactor reflect.Value, input any, e error) {
	doErrorHook := interactor.MethodByName("DoErrorHook")
	if !doErrorHook.IsValid() {
		return
	}
	doErrorHook.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(input), reflect.ValueOf(e)})
}

func (a *API) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(a.m))
	for k := range a.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := strings.Builder{}

	b.WriteString("{")

	for i, d := range a.UseCases() {
		if i > 0 {
			b.WriteString(",")
		}
		ucJSON, _ := json.Marshal(d)
		b.WriteString(fmt.Sprintf("%q: %s", d.Operation(), ucJSON))
	}

	b.WriteString("}")

	return []byte(b.String()), nil
}

type APIBuilder struct {
	api *API
}

func NewAPIBuilder() *APIBuilder {
	return &APIBuilder{
		api: newAPI(),
	}
}

func (b *APIBuilder) WithDescription(desc string) *APIBuilder {
	b.api.description = desc
	return b
}

func (b *APIBuilder) WithHook(hook *GroupHook) *APIBuilder {
	b.api.root.hook = hook
	return b
}

func (b *APIBuilder) AddBeforeHook(hook BeforeHook[any]) *APIBuilder {
	b.api.root.hook.AddBefore(hook)
	return b
}

func (b *APIBuilder) AddAfterHook(hook AfterHook[any, any]) *APIBuilder {
	b.api.root.hook.AddAfter(hook)
	return b
}

func (b *APIBuilder) AddErrorHook(hook ErrorHook[any]) *APIBuilder {
	b.api.root.hook.AddError(hook)
	return b
}

func (b *APIBuilder) AddUseCase(d Descriptor) *APIBuilder {
	op := d.Operation()
	b.api.m[op] = d
	return b
}

func (b *APIBuilder) WithOptions(opts ...APIOptionFunc) *APIBuilder {
	for _, opt := range opts {
		opt(b.api.options)
	}
	return b
}

func (b *APIBuilder) Build() *API {
	return b.api
}

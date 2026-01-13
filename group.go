package grepo

import "context"

type BeforeHook[I any] func(ctx context.Context, desc Descriptor, i I) (context.Context, error)
type AfterHook[I any, O any] func(ctx context.Context, desc Descriptor, i I, o O)
type ErrorHook[I any] func(ctx context.Context, desc Descriptor, i I, e error)

type GroupHook struct {
	before []BeforeHook[any]
	after  []AfterHook[any, any]
	error  []ErrorHook[any]
}

func NewGroupHook() *GroupHook {
	return &GroupHook{}
}

func (h *GroupHook) AddBefore(hook BeforeHook[any]) *GroupHook {
	h.before = append(h.before, hook)
	return h
}

func (h *GroupHook) AddAfter(hook AfterHook[any, any]) *GroupHook {
	h.after = append(h.after, hook)
	return h
}

func (h *GroupHook) AddError(hook ErrorHook[any]) *GroupHook {
	h.error = append(h.error, hook)
	return h
}

type Group struct {
	name string
	hook *GroupHook
}

func NewGroup(name string) *Group {
	return &Group{
		name: name,
		hook: NewGroupHook(),
	}
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) MarshalJSON() ([]byte, error) {
	return []byte(`"` + g.name + `"`), nil
}

func hookBefore(ctx context.Context, desc Descriptor, input any, groups []*Group) (context.Context, error) {
	for _, g := range groups {
		var err error
		for _, beforeHook := range g.hook.before {
			ctx, err = beforeHook(ctx, desc, input)
			if err != nil {
				return ctx, err
			}
		}
	}
	if ctx == nil {
		return nil, ErrNotFound
	}
	return ctx, nil
}

func hookAfter(ctx context.Context, desc Descriptor, input any, output any, groups []*Group) {
	for i := len(groups) - 1; i >= 0; i-- {
		g := groups[i]
		for _, afterHook := range g.hook.after {
			afterHook(ctx, desc, input, output)
		}
	}
}

func hookError(ctx context.Context, desc Descriptor, input any, err error, groups []*Group) {
	for i := len(groups) - 1; i >= 0; i-- {
		g := groups[i]
		for _, errorHook := range g.hook.error {
			errorHook(ctx, desc, input, err)
		}
	}
}

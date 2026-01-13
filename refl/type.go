package refl

import (
	"fmt"
	"reflect"
	"strings"
)

type Field struct {
	Field    string
	Type     *Type
	Optional bool     `json:",omitempty"`
	Enum     []string `json:",omitempty"`
	Custom   []string `json:",omitempty"`
	Min      *int     `json:",omitempty"`
	Max      *int     `json:",omitempty"`
	parent   *Type
}

func (f *Field) Parent() *Type {
	return f.parent
}

type Type struct {
	Kind    string
	Pkg     string `json:",omitempty"`
	Name    string
	Fields  []*Field `json:",omitempty"`
	Element *Type    `json:",omitempty"`
}

func TypeOf(v any) *Type {
	return TypeFor(reflect.TypeOf(v))
}

func TypeFor(t reflect.Type) *Type {
	rt := stripPointer(t)
	s := &Type{
		Kind: kindOf(rt),
		Pkg:  rt.PkgPath(),
		Name: nameOf(t),
	}

	switch s.Kind {
	case KindObject:
		s.Fields = make([]*Field, 0)
		numField := rt.NumField()
		for i := 0; i < numField; i++ {
			ft := rt.Field(i)
			if !ft.IsExported() {
				continue
			}

			f := &Field{
				Field:  ft.Name,
				Type:   TypeFor(ft.Type),
				parent: s,
			}

			tag := ft.Tag.Get("grepo")
			parts := strings.Split(tag, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				kv := strings.SplitN(part, ":", 2)
				if len(kv) != 2 {
					continue
				}
				switch kv[0] {
				case "optional":
					f.Optional = kv[1] == "true"
				case "enum":
					enumValues := strings.Split(kv[1], ",")
					for i := range enumValues {
						enumValues[i] = strings.TrimSpace(enumValues[i])
					}
					f.Enum = enumValues
				case "min", "max":
					value := strings.TrimSpace(kv[1])
					if value != "" {
						var mv int
						fmt.Sscanf(value, "%d", &mv)
						if kv[0] == "max" {
							f.Max = &mv
							continue
						}
						f.Min = &mv
					}

				case "custom":
					customValues := strings.Split(kv[1], ",")
					for i := range customValues {
						customValues[i] = strings.TrimSpace(customValues[i])
					}
					f.Custom = customValues
				}
			}

			s.Fields = append(s.Fields, f)
		}
	case KindArray:
		elemType := TypeFor(rt.Elem())
		s.Element = elemType
	}
	return s
}

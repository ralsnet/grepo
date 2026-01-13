package refl

import (
	"reflect"
	"strings"
)

func nameOf(t reflect.Type) string {
	hasPointer := t.Kind() == reflect.Pointer
	t = stripPointer(t)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return "[]" + nameOf(t.Elem())
	}
	name := t.Name()
	pkg := t.PkgPath()
	parts := strings.Split(pkg, "/")
	lastPart := parts[len(parts)-1]
	if lastPart != "" {
		name = lastPart + "." + name
	}
	if hasPointer {
		return "*" + name
	}
	return name
}

func stripPointer(rt reflect.Type) reflect.Type {
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	return rt
}

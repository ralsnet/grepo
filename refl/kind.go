package refl

import "reflect"

const (
	KindObject  = "object"
	KindArray   = "array"
	KindString  = "string"
	KindInt     = "int"
	KindInt8    = "int8"
	KindInt16   = "int16"
	KindInt32   = "int32"
	KindInt64   = "int64"
	KindUint    = "uint"
	KindUint8   = "uint8"
	KindUint16  = "uint16"
	KindUint32  = "uint32"
	KindUint64  = "uint64"
	KindFloat32 = "float32"
	KindFloat64 = "float64"
	KindBool    = "bool"
	KindTime    = "time"
)

func kindOf(t reflect.Type) string {
	t = stripPointer(t)
	switch t.Kind() {
	case reflect.Struct:
		if t.Name() == "Time" && t.PkgPath() == "time" {
			return KindTime
		}
		return KindObject
	case reflect.Slice, reflect.Array:
		return KindArray
	case reflect.String:
		return KindString
	case reflect.Int:
		return KindInt
	case reflect.Int8:
		return KindInt8
	case reflect.Int16:
		return KindInt16
	case reflect.Int32:
		return KindInt32
	case reflect.Int64:
		return KindInt64
	case reflect.Uint:
		return KindUint
	case reflect.Uint8:
		return KindUint8
	case reflect.Uint16:
		return KindUint16
	case reflect.Uint32:
		return KindUint32
	case reflect.Uint64:
		return KindUint64
	case reflect.Float32:
		return KindFloat32
	case reflect.Float64:
		return KindFloat64
	case reflect.Bool:
		return KindBool
	default:
		return "unknown"
	}
}

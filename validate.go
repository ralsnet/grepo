package grepo

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ralsnet/grepo/refl"
)

type FieldValidator interface {
	Validate(v reflect.Value, f *refl.Field) error
}

type FieldValidatorFunc func(v reflect.Value, f *refl.Field) error

func (fn FieldValidatorFunc) Validate(v reflect.Value, f *refl.Field) error {
	return fn(v, f)
}

func Validate(v any, validators ...FieldValidator) error {
	rv := reflect.ValueOf(v)
	if err := validate(rv, validators...); err != nil {
		return errors.Join(ErrInvalid, err)
	}
	return nil
}

func validate(v reflect.Value, validators ...FieldValidator) error {
	if !v.IsValid() {
		return fmt.Errorf("invalid value")
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	t := refl.TypeOf(v.Interface())
	switch t.Kind {
	case refl.KindObject:
		for _, ft := range t.Fields {
			fv := v.FieldByName(ft.Field)
			if err := validateField(fv, ft, validators...); err != nil {
				return err
			}
		}
	case refl.KindArray:
		for i := 0; i < v.Len(); i++ {
			if err := validate(v.Index(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateField(v reflect.Value, f *refl.Field, validators ...FieldValidator) error {
	if !v.IsValid() {
		return fmt.Errorf("field %s is required but invalid", f.Field)
	}

	rv := v
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			break
		}
		rv = rv.Elem()
	}

	vs := make([]FieldValidator, 0, len(validators)+3)
	vs = append(vs, validators...)
	vs = append(vs, FieldValidatorFunc(validateOptional))
	vs = append(vs, FieldValidatorFunc(validateEnum))
	vs = append(vs, FieldValidatorFunc(validateMinMax))

	for _, validator := range vs {
		if err := validator.Validate(rv, f); err != nil {
			return err
		}
	}

	if err := validate(rv, validators...); err != nil {
		return err
	}

	return nil
}

func validateOptional(v reflect.Value, f *refl.Field) error {
	if f.Optional {
		return nil
	}
	switch {
	case v.CanInt():
		if v.Int() == 0 {
			return nil
		}
	case v.CanUint():
		if v.Uint() == 0 {
			return nil
		}
	case v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64:
		if v.Float() == 0 {
			return nil
		}
	}
	if v.IsZero() {
		return fmt.Errorf("field %s is required but zero", f.Field)
	}
	if (v.Kind() == reflect.Slice || v.Kind() == reflect.Map) && v.Len() == 0 {
		return fmt.Errorf("field %s is required but empty", f.Field)
	}
	return nil
}

func validateEnum(v reflect.Value, f *refl.Field) error {
	if len(f.Enum) == 0 {
		return nil
	}

	if f.Type.Kind == refl.KindObject || f.Type.Kind == refl.KindArray {
		return fmt.Errorf("field %s has enum constraint but is complex type", f.Field)
	}
	switch {
	case v.CanInt():
		val := fmt.Sprintf("%d", v.Int())
		for _, enumVal := range f.Enum {
			if val == enumVal {
				return nil
			}
		}
	case v.CanUint():
		val := fmt.Sprintf("%d", v.Uint())
		for _, enumVal := range f.Enum {
			if val == enumVal {
				return nil
			}
		}
	case v.Kind() == reflect.String:
		val := v.Interface().(string)
		for _, enumVal := range f.Enum {
			if val == enumVal {
				return nil
			}
		}
	}
	return fmt.Errorf("field %s has value %s which is not in enum %v", f.Field, v.String(), f.Enum)
}

func validateMinMax(v reflect.Value, f *refl.Field) error {
	if f.Min != nil {
		switch {
		case v.CanInt():
			if v.Int() < int64(*f.Min) {
				return fmt.Errorf("field %s has value %d which is less than min %d", f.Field, v.Int(), *f.Min)
			}
		case v.CanUint():
			if v.Uint() < uint64(*f.Min) {
				return fmt.Errorf("field %s has value %d which is less than min %d", f.Field, v.Uint(), *f.Min)
			}
		case v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64:
			if v.Float() < float64(*f.Min) {
				return fmt.Errorf("field %s has value %f which is less than min %d", f.Field, v.Float(), *f.Min)
			}
		}
	}
	if f.Max != nil {
		switch {
		case v.CanInt():
			if v.Int() > int64(*f.Max) {
				return fmt.Errorf("field %s has value %d which is greater than max %d", f.Field, v.Int(), *f.Max)
			}
		case v.CanUint():
			if v.Uint() > uint64(*f.Max) {
				return fmt.Errorf("field %s has value %d which is greater than max %d", f.Field, v.Uint(), *f.Max)
			}
		case v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64:
			if v.Float() > float64(*f.Max) {
				return fmt.Errorf("field %s has value %f which is greater than max %d", f.Field, v.Float(), *f.Max)
			}
		}
	}
	return nil
}

package mongoclient

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
)

var (
	ErrNotFoundField = fmt.Errorf("not found field name")
)

func MakeBsonM(patchable interface{}) (bson.M, error) {
	val := reflect.ValueOf(patchable)
	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		val = val.Elem()
	}

	bsonM := bson.M{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		if tag, err := bsoncodec.DefaultStructTagParser(val.Type().Field(i)); err != nil {
			return nil, err
		} else if tag.Skip {
			continue
		} else if tag.OmitEmpty && field.IsZero() || !field.CanInterface() {
			continue
		} else if field.Kind() == reflect.Ptr && !field.IsNil() {
			// unpack pointer if possible
			bsonM[tag.Name] = reflect.Indirect(reflect.ValueOf(field.Interface())).Interface()
		} else if !field.IsZero() {
			// assaign underlying value directly
			bsonM[tag.Name] = field.Interface()
		}
	}

	return bsonM, nil
}

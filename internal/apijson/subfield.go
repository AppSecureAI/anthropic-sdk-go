package apijson

import (
	"github.com/anthropics/anthropic-sdk-go/packages/respjson"
	"reflect"
)

func getSubField(root reflect.Value, index []int, name string) reflect.Value {
	// Validate field name to prevent arbitrary field access via reflection
	if !isValidFieldName(name) {
		return reflect.Value{}
	}
	
	strct := root.FieldByIndex(index[:len(index)-1])
	if !strct.IsValid() {
		panic("couldn't find encapsulating struct for field " + name)
	}
	meta := strct.FieldByName("JSON")
	if !meta.IsValid() {
		return reflect.Value{}
	}
	field := meta.FieldByName(name)
	if !field.IsValid() {
		return reflect.Value{}
	}
	return field
}

// isValidFieldName validates that the field name is safe to use with reflection
func isValidFieldName(name string) bool {
	// Reject empty names
	if len(name) == 0 {
		return false
	}
	
	// Limit field name length to prevent abuse
	if len(name) > 100 {
		return false
	}
	
	// Only allow alphanumeric characters, underscores, and hyphens
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_' || char == '-') {
			return false
		}
	}
	
	// Prevent access to potentially sensitive fields
	// Fields starting with underscore are typically private
	if name[0] == '_' {
		return false
	}
	
	return true
}

func setMetadataSubField(root reflect.Value, index []int, name string, meta Field) {
	target := getSubField(root, index, name)
	if !target.IsValid() {
		return
	}

	if target.Type() == reflect.TypeOf(meta) {
		target.Set(reflect.ValueOf(meta))
	} else if respMeta := meta.toRespField(); target.Type() == reflect.TypeOf(respMeta) {
		target.Set(reflect.ValueOf(respMeta))
	}
}

func setMetadataExtraFields(root reflect.Value, index []int, name string, metaExtras map[string]Field) {
	target := getSubField(root, index, name)
	if !target.IsValid() {
		return
	}

	if target.Type() == reflect.TypeOf(metaExtras) {
		target.Set(reflect.ValueOf(metaExtras))
		return
	}

	newMap := make(map[string]respjson.Field, len(metaExtras))
	if target.Type() == reflect.TypeOf(newMap) {
		for k, v := range metaExtras {
			newMap[k] = v.toRespField()
		}
		target.Set(reflect.ValueOf(newMap))
	}
}

func (f Field) toRespField() respjson.Field {
	if f.IsMissing() {
		return respjson.Field{}
	} else if f.IsNull() {
		return respjson.NewField("null")
	} else if f.IsInvalid() {
		return respjson.NewInvalidField(f.raw)
	} else {
		return respjson.NewField(f.raw)
	}
}
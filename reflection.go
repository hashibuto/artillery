package artillery

import (
	"reflect"
	"strings"

	"github.com/hashibuto/mirage"
)

// Reflect attempts to reflect the data in namespace to the provided object
func Reflect(namespace Namespace, obj any) error {
	lowerToKey := map[string]string{}
	ref := mirage.Reflect(obj, "")
	for _, key := range ref.Keys() {
		lowerToKey[strings.ToLower(key)] = key
	}

	refIo := ref.Io()
	for key, value := range namespace {
		if value == nil {
			continue
		}
		lKey := strings.ToLower(key)

		if objKey, ok := lowerToKey[lKey]; ok {
			info, _ := ref.InfoByName(objKey)
			if info.Kind == reflect.Slice {
				vSlice := value.([]any)
				target, _ := refIo.ValueFromName(objKey)
				switch target.(type) {
				case []string:
					newTarg := make([]string, len(vSlice))
					for i, v := range vSlice {
						newTarg[i] = v.(string)
					}
					value = newTarg
				case []int:
					newTarg := make([]int, len(vSlice))
					for i, v := range vSlice {
						newTarg[i] = v.(int)
					}
					value = newTarg
				case []float64:
					newTarg := make([]float64, len(vSlice))
					for i, v := range vSlice {
						newTarg[i] = v.(float64)
					}
					value = newTarg
				case []bool:
					newTarg := make([]bool, len(vSlice))
					for i, v := range vSlice {
						newTarg[i] = v.(bool)
					}
					value = newTarg
				}
			}

			err := refIo.SetValueByName(objKey, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

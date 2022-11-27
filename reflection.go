package artillery

import (
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
		lKey := strings.ToLower(key)

		if objKey, ok := lowerToKey[lKey]; ok {
			err := refIo.SetValueByName(objKey, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

package main

import (
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/ory/fosite"
	"github.com/stoewer/go-strcase"
)

func main() {
	reflector := new(jsonschema.Reflector)
	reflector.AllowAdditionalProperties = true
	reflector.KeyNamer = strcase.LowerCamelCase

	schema := reflector.Reflect(&fosite.DefaultClient{})
	spec := schema.Definitions["DefaultClient"]

	data, _ := spec.MarshalJSON()
	fmt.Println(string(data))
}

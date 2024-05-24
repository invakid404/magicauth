package main

import (
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/ory/fosite"
)

func main() {
	schema := jsonschema.Reflect(&fosite.DefaultClient{})
	spec := schema.Definitions["DefaultClient"]

	data, _ := spec.MarshalJSON()
	fmt.Println(string(data))
}

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/ory/fosite"
	"github.com/stoewer/go-strcase"
	"log"
	"text/template"
)

//go:embed oauthClient.tmpl
var oauthConfigTemplate string

func main() {
	tmpl, err := template.New("oauthClient").Parse(oauthConfigTemplate)
	if err != nil {
		log.Fatalln("failed to parse template:", err)
	}

	reflector := new(jsonschema.Reflector)
	reflector.AllowAdditionalProperties = true
	reflector.KeyNamer = strcase.LowerCamelCase

	schema := reflector.Reflect(&fosite.DefaultClient{})
	spec := schema.Definitions["DefaultClient"]

	data, _ := spec.MarshalJSON()

	var output bytes.Buffer
	if err = tmpl.Execute(&output, map[string]any{
		"Spec": string(data),
	}); err != nil {
		log.Fatalln("failed to render template:", err)
	}

	fmt.Println(string(output.Bytes()))
}

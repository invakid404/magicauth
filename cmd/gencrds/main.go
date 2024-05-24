package main

import (
	"bytes"
	_ "embed"
	"github.com/invopop/jsonschema"
	"github.com/ory/fosite"
	"github.com/stoewer/go-strcase"
	"github.com/tidwall/pretty"
	"log"
	"os"
	"strings"
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

	// Remove content encoding from secret fields
	for pair := spec.Properties.Oldest(); pair != nil; pair = pair.Next() {
		if strings.Contains(pair.Key, "Secret") {
			pair.Value.ContentEncoding = ""
			if pair.Value.Items != nil {
				pair.Value.Items.ContentEncoding = ""
			}
		}
	}

	// Make id non-required
	writeIdx := 0
	for _, prop := range spec.Required {
		if prop != "id" {
			spec.Required[writeIdx] = prop
			writeIdx++
		}
	}
	spec.Required = spec.Required[:writeIdx]

	data, _ := spec.MarshalJSON()

	var output bytes.Buffer
	if err = tmpl.Execute(&output, map[string]any{
		"Spec": string(data),
	}); err != nil {
		log.Fatalln("failed to render template:", err)
	}

	err = os.WriteFile("./crds/oauthclient.json", pretty.Pretty(output.Bytes()), 0644)
	if err != nil {
		log.Fatalln("failed to write crd:", err)
	}
}

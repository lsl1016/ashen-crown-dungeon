package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ashen-crown-dungeon/internal/game"
)

func main() {
	out := flag.String("out", "./docs/schemas", "output directory")
	flag.Parse()
	if err := os.MkdirAll(filepath.Join(*out, "tools"), 0755); err != nil {
		panic(err)
	}
	defs := game.AgentToolDefinitions()
	catalog := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title":   "Ashen Crown AI GM Tool Catalog",
		"version": "1.0",
		"tools":   defs,
	}
	mustWrite(filepath.Join(*out, "agent-tools.json"), catalog)
	envelope := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id":     "https://ashen-crown.local/schema/agent-tool-execute.json",
		"title":   "AI Tool Gateway Execute Request",
		"type":    "object",
		"properties": map[string]any{
			"tool":      map[string]any{"type": "string", "minLength": 1},
			"arguments": map[string]any{"type": "object"},
		},
		"required":             []string{"tool", "arguments"},
		"additionalProperties": false,
	}
	mustWrite(filepath.Join(*out, "tool-gateway-execute.schema.json"), envelope)
	for _, def := range defs {
		doc := map[string]any{
			"$schema":      "https://json-schema.org/draft/2020-12/schema",
			"name":         def.Name,
			"title":        def.Title,
			"description":  def.Description,
			"inputSchema":  def.InputSchema,
			"outputSchema": def.OutputSchema,
			"annotations":  def.Annotations,
			"_meta":        def.Meta,
		}
		mustWrite(filepath.Join(*out, "tools", def.Name+".json"), doc)
	}
	fmt.Printf("exported %d tools to %s\n", len(defs), *out)
}

func mustWrite(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0644); err != nil {
		panic(err)
	}
}

package vuego

import (
	"encoding/json"
	"fmt"
	"io/fs"

	yaml "gopkg.in/yaml.v3"
)

// readTemplateFile reads filename from the template filesystem. A filename
// that resolves against the context stack is taken as a variable reference, so
// both file("style.css") and file(path) work.
func readTemplateFile(v *Vue, ctx *VueContext, filename string) ([]byte, string, error) {
	if ctx != nil {
		if val, ok := ctx.Stack().Resolve(filename); ok {
			filename = fmt.Sprint(val)
		}
	}

	if v.templateFS == nil {
		return nil, filename, fmt.Errorf("error loading file from nil fs: %s", filename)
	}

	d, err := fs.ReadFile(v.templateFS, filename)
	if err != nil {
		return nil, filename, err
	}
	return d, filename, nil
}

// fileFunc backs the "file" template function, returning file contents as a string.
func fileFunc(v *Vue) func(*VueContext, string) (any, error) {
	return func(ctx *VueContext, filename string) (any, error) {
		d, _, err := readTemplateFile(v, ctx, filename)
		if err != nil {
			return "", err
		}
		return string(d), nil
	}
}

// jsonFileFunc backs the "jsonFile" template function, decoding a JSON object.
func jsonFileFunc(v *Vue) func(*VueContext, string) (any, error) {
	return func(ctx *VueContext, filename string) (any, error) {
		d, name, err := readTemplateFile(v, ctx, filename)
		if err != nil {
			return "", err
		}

		var result map[string]any
		if err := json.Unmarshal(d, &result); err != nil {
			return "", fmt.Errorf("error decoding json from file %s: %w", name, err)
		}
		return result, nil
	}
}

// yamlFileFunc backs the "yamlFile" template function, decoding a YAML mapping.
func yamlFileFunc(v *Vue) func(*VueContext, string) (any, error) {
	return func(ctx *VueContext, filename string) (any, error) {
		d, name, err := readTemplateFile(v, ctx, filename)
		if err != nil {
			return "", err
		}

		var result map[string]any
		if err := yaml.Unmarshal(d, &result); err != nil {
			return "", fmt.Errorf("error decoding yaml from file %s: %w", name, err)
		}
		return result, nil
	}
}

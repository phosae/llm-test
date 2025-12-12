package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

func LoadCaseBody(casesDir, caseName, reqType string) (map[string]interface{}, error) {
	caseFile := filepath.Join(casesDir, caseName, reqType+".json")
	data, err := os.ReadFile(caseFile)
	if err != nil {
		if os.IsNotExist(err) {
			homeDir, homeErr := os.UserHomeDir()
			if homeErr == nil {
				homeCaseFile := filepath.Join(homeDir, ".llm-test", "cases", caseName, reqType+".json")
				data, err = os.ReadFile(homeCaseFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read case file, tried:\n  - %s\n  - %s", caseFile, homeCaseFile)
				}
			} else {
				return nil, fmt.Errorf("failed to read case file %s: %w", caseFile, err)
			}
		} else {
			return nil, fmt.Errorf("failed to read case file %s: %w", caseFile, err)
		}
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse case JSON: %w", err)
	}

	return body, nil
}

func ApplyModelOverride(body map[string]interface{}, model string) {
	if model != "" {
		body["model"] = model
	}
}

func ApplyStreamOverride(body map[string]interface{}, stream bool) {
	if stream {
		body["stream"] = true
	}
}

func ApplyJSONPatch(body map[string]interface{}, patch string) (map[string]interface{}, error) {
	if patch == "" {
		return body, nil
	}

	original, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	merged, err := jsonpatch.MergePatch(original, []byte(patch))
	if err != nil {
		return nil, fmt.Errorf("failed to apply patch: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(merged, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal patched body: %w", err)
	}

	return result, nil
}

func GetModelFromBody(body map[string]interface{}) string {
	if model, ok := body["model"].(string); ok {
		return model
	}
	return ""
}

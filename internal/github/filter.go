package github

import "strings"

func IsFromOrganization(payload map[string]any, organization string) bool {
	repository, ok := payload["repository"].(map[string]any)
	if !ok {
		return false
	}

	owner, ok := repository["owner"].(map[string]any)
	if !ok {
		return false
	}

	login, ok := owner["login"].(string)
	if !ok {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(login), strings.TrimSpace(organization))
}

package github

import "testing"

func TestIsFromOrganizationReturnsTrueForMatchingOwner(t *testing.T) {
	payload := map[string]any{
		"repository": map[string]any{
			"owner": map[string]any{
				"login": "usj",
			},
		},
	}

	if !IsFromOrganization(payload, "usj") {
		t.Fatal("expected matching organization owner to pass")
	}
}

func TestIsFromOrganizationReturnsFalseForDifferentOwner(t *testing.T) {
	payload := map[string]any{
		"repository": map[string]any{
			"owner": map[string]any{
				"login": "someone-else",
			},
		},
	}

	if IsFromOrganization(payload, "usj") {
		t.Fatal("expected non-matching organization owner to fail")
	}
}

func TestIsFromOrganizationReturnsFalseWhenOwnerMissing(t *testing.T) {
	payload := map[string]any{
		"repository": map[string]any{},
	}

	if IsFromOrganization(payload, "usj") {
		t.Fatal("expected malformed payload to fail organization check")
	}
}

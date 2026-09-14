package naming_test

import (
	"testing"

	"github.com/loy-go/loy/internal/generator/naming"
)

func TestCasingConversions(t *testing.T) {
	tests := []struct {
		input       string
		pascal      string
		camel       string
		snake       string
		kebab       string
		packageName string
	}{
		{
			input:       "user_account",
			pascal:      "UserAccount",
			camel:       "userAccount",
			snake:       "user_account",
			kebab:       "user-account",
			packageName: "useraccount",
		},
		{
			input:       "UserAccount",
			pascal:      "UserAccount",
			camel:       "userAccount",
			snake:       "user_account",
			kebab:       "user-account",
			packageName: "useraccount",
		},
		{
			input:       "user-account-detail",
			pascal:      "UserAccountDetail",
			camel:       "userAccountDetail",
			snake:       "user_account_detail",
			kebab:       "user-account-detail",
			packageName: "useraccountdetail",
		},
		{
			input:       "HTTPClient",
			pascal:      "HttpClient",
			camel:       "httpClient",
			snake:       "http_client",
			kebab:       "http-client",
			packageName: "httpclient",
		},
		{
			input:       "user",
			pascal:      "User",
			camel:       "user",
			snake:       "user",
			kebab:       "user",
			packageName: "user",
		},
		{
			input:       "",
			pascal:      "",
			camel:       "",
			snake:       "",
			kebab:       "",
			packageName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := naming.ToPascalCase(tt.input); got != tt.pascal {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.pascal)
			}
			if got := naming.ToCamelCase(tt.input); got != tt.camel {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, got, tt.camel)
			}
			if got := naming.ToSnakeCase(tt.input); got != tt.snake {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.snake)
			}
			if got := naming.ToKebabCase(tt.input); got != tt.kebab {
				t.Errorf("ToKebabCase(%q) = %q, want %q", tt.input, got, tt.kebab)
			}
			if got := naming.ToPackageName(tt.input); got != tt.packageName {
				t.Errorf("ToPackageName(%q) = %q, want %q", tt.input, got, tt.packageName)
			}
		})
	}
}

func TestPluralization(t *testing.T) {
	tests := []struct {
		singular string
		plural   string
	}{
		{"user", "users"},
		{"category", "categories"},
		{"box", "boxes"},
		{"dish", "dishes"},
		{"match", "matches"},
		{"status", "statuses"},
		{"bus", "buses"},
		{"person", "people"},
		{"man", "men"},
		{"datum", "data"},
		{"boy", "boys"},
		{"day", "days"},
		{"User", "Users"},
		{"Category", "Categories"},
		{"Person", "People"},
	}

	for _, tt := range tests {
		t.Run(tt.singular+"->"+tt.plural, func(t *testing.T) {
			if got := naming.Pluralize(tt.singular); got != tt.plural {
				t.Errorf("Pluralize(%q) = %q, want %q", tt.singular, got, tt.plural)
			}
			if got := naming.Singularize(tt.plural); got != tt.singular {
				t.Errorf("Singularize(%q) = %q, want %q", tt.plural, got, tt.singular)
			}
		})
	}

	if got := naming.Pluralize(""); got != "" {
		t.Errorf("Pluralize empty = %q, want empty", got)
	}
	if got := naming.Singularize(""); got != "" {
		t.Errorf("Singularize empty = %q, want empty", got)
	}
}

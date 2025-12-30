package gemini_test

import (
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai/gemini"
)

func TestNewProvider_CreatesProvider(t *testing.T) {
	provider := gemini.NewProvider("test-api-key")

	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

func TestSetModel_UpdatesModelAndReturnsProvider(t *testing.T) {
	provider := gemini.NewProvider("test-api-key")
	result := provider.SetModel("gemini-pro")

	if result == nil {
		t.Error("expected SetModel to return the provider for chaining")
	}
}

func TestSetTemperature_SetsTemperatureAndReturnsProvider(t *testing.T) {
	provider := gemini.NewProvider("test-api-key")
	result := provider.SetTemperature(0.8)

	if result == nil {
		t.Error("expected SetTemperature to return the provider for chaining")
	}
}

func TestSetMaxTokens_SetsMaxTokensAndReturnsProvider(t *testing.T) {
	provider := gemini.NewProvider("test-api-key")
	result := provider.SetMaxTokens(2048)

	if result == nil {
		t.Error("expected SetMaxTokens to return the provider for chaining")
	}
}

func TestSetBaseURL_UpdatesBaseURLAndReturnsProvider(t *testing.T) {
	provider := gemini.NewProvider("test-api-key")
	customURL := "https://custom.gemini.api.com/v1"
	result := provider.SetBaseURL(customURL)

	if result == nil {
		t.Error("expected SetBaseURL to return the provider for chaining")
	}
}

func TestProvider_MethodChainingWorks(t *testing.T) {
	provider := gemini.NewProvider("test-api-key").
		SetModel("gemini-pro").
		SetTemperature(0.5).
		SetMaxTokens(1024)

	if provider == nil {
		t.Error("expected method chaining to work")
	}
}

package openai_test

import (
	"testing"

	"github.com/marti-jorda-roca/gopher-ai/gopherai/openai"
)

func TestNewProvider_SetsDefaultModel(t *testing.T) {
	provider := openai.NewProvider("test-api-key")

	if provider == nil {
		t.Fatal("expected provider to be created")
	}
}

func TestSetModel_UpdatesModelAndReturnsProvider(t *testing.T) {
	provider := openai.NewProvider("test-api-key")
	result := provider.SetModel("gpt-4o")

	if result == nil {
		t.Error("expected SetModel to return the provider for chaining")
	}
}

func TestSetTemperature_SetsTemperatureAndReturnsProvider(t *testing.T) {
	provider := openai.NewProvider("test-api-key")
	result := provider.SetTemperature(0.7)

	if result == nil {
		t.Error("expected SetTemperature to return the provider for chaining")
	}
}

func TestSetMaxTokens_SetsMaxTokensAndReturnsProvider(t *testing.T) {
	provider := openai.NewProvider("test-api-key")
	result := provider.SetMaxTokens(1000)

	if result == nil {
		t.Error("expected SetMaxTokens to return the provider for chaining")
	}
}

func TestSetBaseURL_UpdatesBaseURLAndReturnsProvider(t *testing.T) {
	provider := openai.NewProvider("test-api-key")
	customURL := "https://custom.api.com/v1"
	result := provider.SetBaseURL(customURL)

	if result == nil {
		t.Error("expected SetBaseURL to return the provider for chaining")
	}
}

func TestProvider_MethodChainingWorks(t *testing.T) {
	provider := openai.NewProvider("test-api-key").
		SetModel("gpt-4o").
		SetTemperature(0.5).
		SetMaxTokens(2000)

	if provider == nil {
		t.Error("expected method chaining to work")
	}
}

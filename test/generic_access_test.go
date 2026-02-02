package test

import (
	"testing"

	config233 "github.com/neko233-com/config233-go/pkg/config233"
)

func TestGenericAccess_ItemConfig(t *testing.T) {
	// Use a dedicated manager pointed at testdata and set it as the global Instance
	manager := config233.NewConfigManager233("../testdata")
	config233.Instance = manager
	config233.RegisterType[ItemConfig]()

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("load configs failed: %v", err)
	}

	cfg, ok := config233.GetConfigById[ItemConfig]("1")
	if !ok || cfg == nil {
		t.Fatalf("expected config id=1, got nil")
	}
	if cfg.Itemid != 1 {
		t.Fatalf("expected Itemid=1, got %d", cfg.Itemid)
	}

	list := config233.GetConfigList[ItemConfig]()
	if len(list) == 0 {
		t.Fatalf("expected non-empty ItemConfig list")
	}
}

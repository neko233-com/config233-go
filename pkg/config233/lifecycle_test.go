package config233

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// LifecycleConfig 测试生命周期接口
type LifecycleConfig struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// unexported field to verify AfterLoad
	processed bool
}

func (c *LifecycleConfig) AfterLoad() {
	c.Name = c.Name + "_Processed"
	c.processed = true
}

func (c *LifecycleConfig) Check() error {
	return nil
}

// ValidatorConfig 测试校验接口
type ValidatorConfig struct {
	Id         string `json:"id"`
	ShouldFail bool   `json:"shouldFail"`
}

func (c *ValidatorConfig) Check() error {
	if c.ShouldFail {
		return fmt.Errorf("Validation failed as expected")
	}
	return nil
}

func TestLifecycleAndValidator(t *testing.T) {
	// 确保单例被重置 (Simple hack: tests run sequentially usually)

	t.Run("Lifecycle", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "LifecycleConfig.json")

		data := []map[string]interface{}{
			{"id": "1", "name": "Test1"},
		}
		fileContent, _ := json.Marshal(data)
		os.WriteFile(configFile, fileContent, 0644)

		manager := NewConfigManager233(tempDir)
		manager.RegisterType(reflect.TypeOf((*LifecycleConfig)(nil)).Elem())

		if err := manager.loadJsonConfig(configFile); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Verify using GetConfigMap
		loadedMap := GetConfigMap[LifecycleConfig]()
		if loadedMap == nil {
			t.Fatal("Config map is nil")
		}

		cfg, ok := loadedMap["1"]
		if !ok {
			t.Fatal("Config id 1 not found")
		}

		if cfg.Name != "Test1_Processed" {
			t.Errorf("Expected Name 'Test1_Processed', got '%s'", cfg.Name)
		}
		if !cfg.processed {
			t.Error("processed should be true")
		}
	})

	t.Run("Validator", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "ValidatorConfig.json")

		data := []map[string]interface{}{
			{"id": "1", "shouldFail": false},
			{"id": "2", "shouldFail": true},
		}
		fileContent, _ := json.Marshal(data)
		os.WriteFile(configFile, fileContent, 0644)

		manager := NewConfigManager233(tempDir)
		manager.RegisterType(reflect.TypeOf((*ValidatorConfig)(nil)).Elem())

		if err := manager.loadJsonConfig(configFile); err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Verify map
		loadedMap := GetConfigMap[ValidatorConfig]()

		if _, ok := loadedMap["1"]; !ok {
			t.Error("Item 1 should be loaded")
		}
		if _, ok := loadedMap["2"]; ok {
			t.Error("Item 2 should NOT be loaded (validation failed)")
		}

		// Verify list
		loadedList := GetConfigList[ValidatorConfig]()
		found1 := false
		found2 := false
		for _, v := range loadedList {
			if v.Id == "1" {
				found1 = true
			}
			if v.Id == "2" {
				found2 = true
			}
		}
		if !found1 {
			t.Error("Item 1 should be in list")
		}
		if found2 {
			t.Error("Item 2 should NOT be in list")
		}
	})
}

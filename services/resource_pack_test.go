package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"minecraft-easyserver/models"
)

func TestResourcePackService(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "bedrock_test")
	if err != nil {
		t.Fatal("Failed to create temporary directory:", err)
	}
	defer os.RemoveAll(tempDir)

	// Set bedrock path
	SetBedrockPath(tempDir)

	// Create test server.properties first
	serverPropsPath := filepath.Join(tempDir, "server.properties")
	serverPropsContent := "level-name=test_world\n"
	if err := os.WriteFile(serverPropsPath, []byte(serverPropsContent), 0644); err != nil {
		t.Fatal("Failed to write server.properties:", err)
	}

	// Create test world directory
	worldsPath := filepath.Join(tempDir, "worlds", "test_world")
	if err := os.MkdirAll(worldsPath, 0755); err != nil {
		t.Fatal("Failed to create world directory:", err)
	}

	// Create resource pack service
	service := NewResourcePackService()

	// Test GetResourcePacks with empty directory
	t.Run("GetResourcePacks_EmptyDirectory", func(t *testing.T) {
		packs, err := service.GetResourcePacks()
		if err != nil {
			t.Fatal("Failed to get resource packs:", err)
		}
		if len(packs) != 0 {
			t.Errorf("Expected 0 resource packs, got %d", len(packs))
		}
	})

	// Create test resource pack directory structure
	resourcePacksPath := filepath.Join(tempDir, "resource_packs")
	testPackPath := filepath.Join(resourcePacksPath, "test_pack")
	if err := os.MkdirAll(testPackPath, 0755); err != nil {
		t.Fatal("Failed to create test pack directory:", err)
	}

	// Create test manifest.json
	manifest := models.ResourcePackManifest{
		FormatVersion: 2,
		Header: models.ResourcePackHeader{
			Description: "Test Resource Pack",
			Name:        "Test Pack",
			UUID:        "12345678-1234-1234-1234-123456789012",
			Version:     [3]int{1, 0, 0},
		},
		Modules: []models.ResourcePackModule{
			{
				Description: "Test Module",
				Type:        "resources",
				UUID:        "87654321-4321-4321-4321-210987654321",
				Version:     [3]int{1, 0, 0},
			},
		},
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal("Failed to marshal manifest:", err)
	}

	manifestPath := filepath.Join(testPackPath, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		t.Fatal("Failed to write manifest file:", err)
	}

	// Test GetResourcePacks with test pack
	t.Run("GetResourcePacks_WithTestPack", func(t *testing.T) {
		packs, err := service.GetResourcePacks()
		if err != nil {
			t.Fatal("Failed to get resource packs:", err)
		}
		if len(packs) != 1 {
			t.Errorf("Expected 1 resource pack, got %d", len(packs))
		}
		if packs[0].Name != "Test Pack" {
			t.Errorf("Expected pack name 'Test Pack', got '%s'", packs[0].Name)
		}
		if packs[0].UUID != "12345678-1234-1234-1234-123456789012" {
			t.Errorf("Expected pack UUID '12345678-1234-1234-1234-123456789012', got '%s'", packs[0].UUID)
		}
		if packs[0].Active {
			t.Error("Expected pack to be inactive initially")
		}
	})

	// Test ActivateResourcePack
	t.Run("ActivateResourcePack", func(t *testing.T) {
		err := service.ActivateResourcePack("12345678-1234-1234-1234-123456789012")
		if err != nil {
			t.Fatal("Failed to activate resource pack:", err)
		}

		// Check if world_resource_packs.json was created
		worldResourcePacksPath := filepath.Join(worldsPath, "world_resource_packs.json")
		if _, err := os.Stat(worldResourcePacksPath); os.IsNotExist(err) {
			t.Error("world_resource_packs.json was not created")
		}

		// Read and verify content
		data, err := os.ReadFile(worldResourcePacksPath)
		if err != nil {
			t.Fatal("Failed to read world_resource_packs.json:", err)
		}

		var activePacks []models.WorldResourcePack
		if err := json.Unmarshal(data, &activePacks); err != nil {
			t.Fatal("Failed to unmarshal world_resource_packs.json:", err)
		}

		if len(activePacks) != 1 {
			t.Errorf("Expected 1 active pack, got %d", len(activePacks))
		}
		if activePacks[0].PackID != "12345678-1234-1234-1234-123456789012" {
			t.Errorf("Expected pack ID '12345678-1234-1234-1234-123456789012', got '%s'", activePacks[0].PackID)
		}
	})

	// Test GetResourcePacks after activation
	t.Run("GetResourcePacks_AfterActivation", func(t *testing.T) {
		packs, err := service.GetResourcePacks()
		if err != nil {
			t.Fatal("Failed to get resource packs:", err)
		}
		if len(packs) != 1 {
			t.Errorf("Expected 1 resource pack, got %d", len(packs))
		}
		if !packs[0].Active {
			t.Error("Expected pack to be active after activation")
		}
	})

	// Test ActivateResourcePack with already activated pack
	t.Run("ActivateResourcePack_AlreadyActivated", func(t *testing.T) {
		err := service.ActivateResourcePack("12345678-1234-1234-1234-123456789012")
		if err == nil {
			t.Error("Expected error when activating already activated pack")
		}
	})

	// Test DeactivateResourcePack
	t.Run("DeactivateResourcePack", func(t *testing.T) {
		err := service.DeactivateResourcePack("12345678-1234-1234-1234-123456789012")
		if err != nil {
			t.Fatal("Failed to deactivate resource pack:", err)
		}

		// Check if pack is removed from world_resource_packs.json
		worldResourcePacksPath := filepath.Join(worldsPath, "world_resource_packs.json")
		data, err := os.ReadFile(worldResourcePacksPath)
		if err != nil {
			t.Fatal("Failed to read world_resource_packs.json:", err)
		}

		var activePacks []models.WorldResourcePack
		if err := json.Unmarshal(data, &activePacks); err != nil {
			t.Fatal("Failed to unmarshal world_resource_packs.json:", err)
		}

		if len(activePacks) != 0 {
			t.Errorf("Expected 0 active packs, got %d", len(activePacks))
		}
	})

	// Test DeactivateResourcePack with not activated pack
	t.Run("DeactivateResourcePack_NotActivated", func(t *testing.T) {
		err := service.DeactivateResourcePack("12345678-1234-1234-1234-123456789012")
		if err == nil {
			t.Error("Expected error when deactivating not activated pack")
		}
	})

	// Test DeleteResourcePack
	t.Run("DeleteResourcePack", func(t *testing.T) {
		err := service.DeleteResourcePack("12345678-1234-1234-1234-123456789012")
		if err != nil {
			t.Fatal("Failed to delete resource pack:", err)
		}

		// Check if directory was deleted
		if _, err := os.Stat(testPackPath); !os.IsNotExist(err) {
			t.Error("Resource pack directory was not deleted")
		}
	})

	// Test GetResourcePacks after deletion
	t.Run("GetResourcePacks_AfterDeletion", func(t *testing.T) {
		packs, err := service.GetResourcePacks()
		if err != nil {
			t.Fatal("Failed to get resource packs:", err)
		}
		if len(packs) != 0 {
			t.Errorf("Expected 0 resource packs after deletion, got %d", len(packs))
		}
	})

	// Test with non-existent pack
	t.Run("ActivateResourcePack_NotFound", func(t *testing.T) {
		err := service.ActivateResourcePack("non-existent-uuid")
		if err == nil {
			t.Error("Expected error when activating non-existent pack")
		}
	})

	t.Run("DeleteResourcePack_NotFound", func(t *testing.T) {
		err := service.DeleteResourcePack("non-existent-uuid")
		if err == nil {
			t.Error("Expected error when deleting non-existent pack")
		}
	})

	// Test official preset pack filtering
	t.Run("OfficialPresetPacks_Filtered", func(t *testing.T) {
		// Create official preset pack directories
		officialPacks := []string{"chemistry", "editor", "vanilla"}
		for _, packName := range officialPacks {
			officialPackPath := filepath.Join(resourcePacksPath, packName)
			if err := os.MkdirAll(officialPackPath, 0755); err != nil {
				t.Fatal("Failed to create official pack directory:", err)
			}

			// Create manifest for official pack
			officialManifest := models.ResourcePackManifest{
				FormatVersion: 2,
				Header: models.ResourcePackHeader{
					Description: "Official " + packName + " Pack",
					Name:        packName,
					UUID:        "official-" + packName + "-uuid",
					Version:     [3]int{1, 0, 0},
				},
			}

			manifestData, err := json.MarshalIndent(officialManifest, "", "  ")
			if err != nil {
				t.Fatal("Failed to marshal official manifest:", err)
			}

			manifestPath := filepath.Join(officialPackPath, "manifest.json")
			if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
				t.Fatal("Failed to write official manifest file:", err)
			}
		}

		// Test that official packs are filtered out
		packs, err := service.GetResourcePacks()
		if err != nil {
			t.Fatal("Failed to get resource packs:", err)
		}

		// Should only return 0 packs (official packs should be filtered out)
		if len(packs) != 0 {
			t.Errorf("Expected 0 resource packs (official packs filtered), got %d", len(packs))
		}

		// Test that official packs cannot be activated
		for _, packName := range officialPacks {
			err := service.ActivateResourcePack("official-" + packName + "-uuid")
			if err == nil {
				t.Errorf("Expected error when trying to activate official pack %s", packName)
			}
		}

		// Test that official packs cannot be deleted
		for _, packName := range officialPacks {
			err := service.DeleteResourcePack("official-" + packName + "-uuid")
			if err == nil {
				t.Errorf("Expected error when trying to delete official pack %s", packName)
			}
		}
	})
}

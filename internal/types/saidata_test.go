package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSoftwareDataFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		wantErr  bool
		validate func(t *testing.T, data *SoftwareData)
	}{
		{
			name: "minimal valid saidata",
			yamlData: `
version: "0.2"
metadata:
  name: "test-software"
  description: "Test software"
`,
			wantErr: false,
			validate: func(t *testing.T, data *SoftwareData) {
				assert.Equal(t, "0.2", data.Version)
				assert.Equal(t, "test-software", data.Metadata.Name)
				assert.Equal(t, "Test software", data.Metadata.Description)
			},
		},
		{
			name: "complete saidata structure",
			yamlData: `
version: "0.2"
metadata:
  name: "apache"
  display_name: "Apache HTTP Server"
  description: "Web server"
  version: "2.4.58"
  category: "web-server"
  tags: ["web", "server"]
  license: "Apache-2.0"
  urls:
    website: "https://httpd.apache.org"
    documentation: "https://httpd.apache.org/docs"
packages:
  - name: "apache2"
    version: "2.4.58"
    alternatives: ["httpd"]
services:
  - name: "apache"
    service_name: "apache2"
    type: "systemd"
    enabled: true
files:
  - name: "config"
    path: "/etc/apache2/apache2.conf"
    type: "config"
    owner: "root"
    mode: "0644"
directories:
  - name: "config"
    path: "/etc/apache2"
    owner: "root"
    mode: "0755"
commands:
  - name: "apache2"
    path: "/usr/sbin/apache2"
    shell_completion: true
ports:
  - port: 80
    protocol: "tcp"
    service: "http"
containers:
  - name: "apache-httpd"
    image: "httpd"
    tag: "2.4"
    ports: ["80:80"]
providers:
  apt:
    packages:
      - name: "apache2"
        version: "2.4.58-1"
`,
			wantErr: false,
			validate: func(t *testing.T, data *SoftwareData) {
				assert.Equal(t, "apache", data.Metadata.Name)
				assert.Len(t, data.Packages, 1)
				assert.Len(t, data.Services, 1)
				assert.Len(t, data.Files, 1)
				assert.Len(t, data.Directories, 1)
				assert.Len(t, data.Commands, 1)
				assert.Len(t, data.Ports, 1)
				assert.Len(t, data.Containers, 1)
				assert.Contains(t, data.Providers, "apt")
				
				// Test default values
				assert.Equal(t, "apache2", data.Services[0].GetServiceNameOrDefault())
				assert.Equal(t, "tcp", data.Ports[0].GetProtocolOrDefault())
			},
		},
		{
			name: "saidata with compatibility matrix",
			yamlData: `
version: "0.2"
metadata:
  name: "test"
compatibility:
  matrix:
    - provider: "apt"
      platform: ["linux"]
      architecture: "amd64"
      os: "ubuntu"
      os_version: ["20.04", "22.04"]
      supported: true
      recommended: true
  versions:
    latest: "1.0.0"
    minimum: "0.9.0"
`,
			wantErr: false,
			validate: func(t *testing.T, data *SoftwareData) {
				require.NotNil(t, data.Compatibility)
				assert.Len(t, data.Compatibility.Matrix, 1)
				
				entry := data.Compatibility.Matrix[0]
				assert.Equal(t, "apt", entry.Provider)
				assert.True(t, entry.Supported)
				assert.True(t, entry.Recommended)
				
				platforms := entry.GetPlatformsAsStrings()
				assert.Equal(t, []string{"linux"}, platforms)
				
				versions := entry.GetOSVersionsAsStrings()
				assert.Equal(t, []string{"20.04", "22.04"}, versions)
				
				assert.Equal(t, "1.0.0", data.Compatibility.Versions.Latest)
			},
		},
		{
			name: "invalid YAML",
			yamlData: `
version: "0.2"
metadata:
  name: "test"
packages:
  - invalid structure
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := LoadSoftwareDataFromYAML([]byte(tt.yamlData))
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, data)
			
			if tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestLoadExistingSaidataFiles(t *testing.T) {
	// Test loading actual saidata files from the samples directory
	saidataFiles := []string{
		"../../docs/saidata_samples/ap/apache/default.yaml",
		"../../docs/saidata_samples/el/elasticsearch/default.yaml",
		"../../docs/saidata_samples/do/docker/default.yaml",
	}

	for _, file := range saidataFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Check if file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Skipf("Saidata file %s does not exist", file)
				return
			}

			data, err := os.ReadFile(file)
			require.NoError(t, err)

			saidata, err := LoadSoftwareDataFromYAML(data)
			require.NoError(t, err)
			require.NotNil(t, saidata)

			// Basic validation
			assert.NotEmpty(t, saidata.Version)
			assert.NotEmpty(t, saidata.Metadata.Name)

			// Validate JSON conversion (for schema validation)
			jsonData, err := saidata.ToJSON()
			require.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Ensure JSON is valid
			var jsonObj map[string]interface{}
			err = json.Unmarshal(jsonData, &jsonObj)
			require.NoError(t, err)
		})
	}
}

func TestSoftwareDataGetters(t *testing.T) {
	saidata := &SoftwareData{
		Packages: []Package{
			{Name: "pkg1", Version: "1.0"},
			{Name: "pkg2", Version: "2.0"},
		},
		Services: []Service{
			{Name: "svc1", ServiceName: "service1"},
			{Name: "svc2"},
		},
		Files: []File{
			{Name: "config", Path: "/etc/config"},
			{Name: "log", Path: "/var/log/app.log"},
		},
		Directories: []Directory{
			{Name: "data", Path: "/var/lib/app"},
		},
		Commands: []Command{
			{Name: "app", Path: "/usr/bin/app"},
		},
		Ports: []Port{
			{Port: 8080, Protocol: "tcp"},
			{Port: 9090, Protocol: "udp"},
		},
		Containers: []Container{
			{Name: "app-container", Image: "app", Tag: "latest"},
		},
		Providers: map[string]ProviderConfig{
			"apt": {
				Packages: []Package{{Name: "app-deb"}},
			},
		},
	}

	t.Run("GetPackageByName", func(t *testing.T) {
		pkg := saidata.GetPackageByName("pkg1")
		require.NotNil(t, pkg)
		assert.Equal(t, "1.0", pkg.Version)

		pkg = saidata.GetPackageByName("nonexistent")
		assert.Nil(t, pkg)
	})

	t.Run("GetServiceByName", func(t *testing.T) {
		svc := saidata.GetServiceByName("svc1")
		require.NotNil(t, svc)
		assert.Equal(t, "service1", svc.ServiceName)

		svc = saidata.GetServiceByName("nonexistent")
		assert.Nil(t, svc)
	})

	t.Run("GetFileByName", func(t *testing.T) {
		file := saidata.GetFileByName("config")
		require.NotNil(t, file)
		assert.Equal(t, "/etc/config", file.Path)

		file = saidata.GetFileByName("nonexistent")
		assert.Nil(t, file)
	})

	t.Run("GetDirectoryByName", func(t *testing.T) {
		dir := saidata.GetDirectoryByName("data")
		require.NotNil(t, dir)
		assert.Equal(t, "/var/lib/app", dir.Path)

		dir = saidata.GetDirectoryByName("nonexistent")
		assert.Nil(t, dir)
	})

	t.Run("GetCommandByName", func(t *testing.T) {
		cmd := saidata.GetCommandByName("app")
		require.NotNil(t, cmd)
		assert.Equal(t, "/usr/bin/app", cmd.Path)

		cmd = saidata.GetCommandByName("nonexistent")
		assert.Nil(t, cmd)
	})

	t.Run("GetPortByNumber", func(t *testing.T) {
		port := saidata.GetPortByNumber(8080)
		require.NotNil(t, port)
		assert.Equal(t, "tcp", port.Protocol)

		port = saidata.GetPortByNumber(9999)
		assert.Nil(t, port)
	})

	t.Run("GetContainerByName", func(t *testing.T) {
		container := saidata.GetContainerByName("app-container")
		require.NotNil(t, container)
		assert.Equal(t, "app", container.Image)

		container = saidata.GetContainerByName("nonexistent")
		assert.Nil(t, container)
	})

	t.Run("GetProviderConfig", func(t *testing.T) {
		config := saidata.GetProviderConfig("apt")
		require.NotNil(t, config)
		assert.Len(t, config.Packages, 1)

		config = saidata.GetProviderConfig("nonexistent")
		assert.Nil(t, config)
	})
}

func TestCompatibilityEntry(t *testing.T) {
	t.Run("GetPlatformsAsStrings", func(t *testing.T) {
		// Test string platform
		entry := CompatibilityEntry{Platform: "linux"}
		platforms := entry.GetPlatformsAsStrings()
		assert.Equal(t, []string{"linux"}, platforms)

		// Test string slice platform
		entry = CompatibilityEntry{Platform: []string{"linux", "macos"}}
		platforms = entry.GetPlatformsAsStrings()
		assert.Equal(t, []string{"linux", "macos"}, platforms)

		// Test interface slice platform
		entry = CompatibilityEntry{Platform: []interface{}{"linux", "macos"}}
		platforms = entry.GetPlatformsAsStrings()
		assert.Equal(t, []string{"linux", "macos"}, platforms)

		// Test nil platform
		entry = CompatibilityEntry{Platform: nil}
		platforms = entry.GetPlatformsAsStrings()
		assert.Nil(t, platforms)
	})

	t.Run("GetOSVersionsAsStrings", func(t *testing.T) {
		entry := CompatibilityEntry{OSVersion: []string{"20.04", "22.04"}}
		versions := entry.GetOSVersionsAsStrings()
		assert.Equal(t, []string{"20.04", "22.04"}, versions)
	})
}

func TestContainerMethods(t *testing.T) {
	t.Run("GetFullImageName", func(t *testing.T) {
		// Test with registry and tag
		container := Container{
			Image:    "nginx",
			Tag:      "1.21",
			Registry: "docker.io",
		}
		assert.Equal(t, "docker.io/nginx:1.21", container.GetFullImageName())

		// Test with only image
		container = Container{Image: "nginx"}
		assert.Equal(t, "nginx", container.GetFullImageName())

		// Test with image and tag
		container = Container{Image: "nginx", Tag: "latest"}
		assert.Equal(t, "nginx:latest", container.GetFullImageName())

		// Test with registry only
		container = Container{Image: "nginx", Registry: "docker.io"}
		assert.Equal(t, "docker.io/nginx", container.GetFullImageName())
	})
}

func TestServiceMethods(t *testing.T) {
	t.Run("GetServiceNameOrDefault", func(t *testing.T) {
		// Test with explicit service name
		service := Service{Name: "apache", ServiceName: "apache2"}
		assert.Equal(t, "apache2", service.GetServiceNameOrDefault())

		// Test with default to name
		service = Service{Name: "nginx"}
		assert.Equal(t, "nginx", service.GetServiceNameOrDefault())
	})
}

func TestCommandMethods(t *testing.T) {
	t.Run("GetPathOrDefault", func(t *testing.T) {
		// Test with explicit path
		command := Command{Name: "docker", Path: "/usr/local/bin/docker"}
		assert.Equal(t, "/usr/local/bin/docker", command.GetPathOrDefault())

		// Test with default path
		command = Command{Name: "docker"}
		assert.Equal(t, "/usr/bin/docker", command.GetPathOrDefault())
	})
}

func TestPortMethods(t *testing.T) {
	t.Run("GetProtocolOrDefault", func(t *testing.T) {
		// Test with explicit protocol
		port := Port{Port: 80, Protocol: "tcp"}
		assert.Equal(t, "tcp", port.GetProtocolOrDefault())

		// Test with default protocol
		port = Port{Port: 80}
		assert.Equal(t, "tcp", port.GetProtocolOrDefault())
	})
}

func TestSoftwareDataToJSON(t *testing.T) {
	saidata := &SoftwareData{
		Version: "0.2",
		Metadata: Metadata{
			Name:        "test",
			Description: "Test software",
		},
		Packages: []Package{
			{Name: "test-pkg", Version: "1.0"},
		},
	}

	jsonData, err := saidata.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	assert.Equal(t, "0.2", result["version"])
	assert.Contains(t, result, "metadata")
	assert.Contains(t, result, "packages")
}

func TestDefaultValues(t *testing.T) {
	yamlData := `
version: "0.2"
metadata:
  name: "test"
services:
  - name: "test-service"
commands:
  - name: "test-cmd"
ports:
  - port: 8080
`

	saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
	require.NoError(t, err)

	// Test default service name
	assert.Equal(t, "test-service", saidata.Services[0].ServiceName)

	// Test default command path
	assert.Equal(t, "/usr/bin/test-cmd", saidata.Commands[0].Path)

	// Test default port protocol
	assert.Equal(t, "tcp", saidata.Ports[0].Protocol)
}
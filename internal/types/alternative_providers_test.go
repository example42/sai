package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceType(t *testing.T) {
	t.Run("complete source configuration", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "nginx"
sources:
  - name: "main"
    url: "http://nginx.org/download/nginx-1.24.0.tar.gz"
    version: "1.24.0"
    build_system: "autotools"
    build_dir: "/tmp/sai-build-nginx"
    source_dir: "/tmp/sai-build-nginx/nginx-1.24.0"
    install_prefix: "/usr/local"
    configure_args:
      - "--with-http_ssl_module"
      - "--with-http_v2_module"
    build_args:
      - "-j$(nproc)"
    install_args:
      - "install"
    prerequisites:
      - "build-essential"
      - "libssl-dev"
    environment:
      CC: "gcc"
      CFLAGS: "-O2 -g"
    checksum: "sha256:5d0b0e8f7e8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f"
    custom_commands:
      download: "wget -O nginx-1.24.0.tar.gz {{url}}"
      extract: "tar -xzf nginx-1.24.0.tar.gz"
      configure: "./configure {{configure_args | join(' ')}}"
      build: "make {{build_args | join(' ')}}"
      install: "make {{install_args | join(' ')}}"
      uninstall: "rm -rf /usr/local/sbin/nginx"
      validation: "nginx -t && nginx -v"
      version: "nginx -v 2>&1 | grep -o 'nginx/[0-9.]*'"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)
		require.Len(t, saidata.Sources, 1)

		source := saidata.Sources[0]
		assert.Equal(t, "main", source.Name)
		assert.Equal(t, "http://nginx.org/download/nginx-1.24.0.tar.gz", source.URL)
		assert.Equal(t, "1.24.0", source.Version)
		assert.Equal(t, "autotools", source.BuildSystem)
		assert.Equal(t, "/tmp/sai-build-nginx", source.BuildDir)
		assert.Equal(t, "/tmp/sai-build-nginx/nginx-1.24.0", source.SourceDir)
		assert.Equal(t, "/usr/local", source.InstallPrefix)
		assert.Len(t, source.ConfigureArgs, 2)
		assert.Contains(t, source.ConfigureArgs, "--with-http_ssl_module")
		assert.Len(t, source.BuildArgs, 1)
		assert.Equal(t, "-j$(nproc)", source.BuildArgs[0])
		assert.Len(t, source.InstallArgs, 1)
		assert.Equal(t, "install", source.InstallArgs[0])
		assert.Len(t, source.Prerequisites, 2)
		assert.Contains(t, source.Prerequisites, "build-essential")
		assert.Equal(t, "gcc", source.Environment["CC"])
		assert.Equal(t, "-O2 -g", source.Environment["CFLAGS"])
		assert.Equal(t, "sha256:5d0b0e8f7e8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f", source.Checksum)
		
		require.NotNil(t, source.CustomCommands)
		assert.Equal(t, "wget -O nginx-1.24.0.tar.gz {{url}}", source.CustomCommands.Download)
		assert.Equal(t, "tar -xzf nginx-1.24.0.tar.gz", source.CustomCommands.Extract)
		assert.Equal(t, "./configure {{configure_args | join(' ')}}", source.CustomCommands.Configure)
		assert.Equal(t, "make {{build_args | join(' ')}}", source.CustomCommands.Build)
		assert.Equal(t, "make {{install_args | join(' ')}}", source.CustomCommands.Install)
		assert.Equal(t, "rm -rf /usr/local/sbin/nginx", source.CustomCommands.Uninstall)
		assert.Equal(t, "nginx -t && nginx -v", source.CustomCommands.Validation)
		assert.Equal(t, "nginx -v 2>&1 | grep -o 'nginx/[0-9.]*'", source.CustomCommands.Version)
	})

	t.Run("minimal source configuration", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "simple"
sources:
  - name: "main"
    url: "https://example.com/source.tar.gz"
    build_system: "make"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)
		require.Len(t, saidata.Sources, 1)

		source := saidata.Sources[0]
		assert.Equal(t, "main", source.Name)
		assert.Equal(t, "https://example.com/source.tar.gz", source.URL)
		assert.Equal(t, "make", source.BuildSystem)
		assert.Empty(t, source.Version)
		assert.Empty(t, source.BuildDir)
		assert.Empty(t, source.SourceDir)
		assert.Empty(t, source.InstallPrefix)
		assert.Nil(t, source.CustomCommands)
	})
}

func TestBinaryType(t *testing.T) {
	t.Run("complete binary configuration", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "terraform"
binaries:
  - name: "main"
    url: "https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_amd64.zip"
    version: "1.6.6"
    architecture: "amd64"
    platform: "linux"
    checksum: "sha256:b8a3892c58c33ee2b4b8e7c2c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8"
    install_path: "/usr/local/bin"
    executable: "terraform"
    permissions: "0755"
    archive:
      format: "zip"
      strip_prefix: ""
      extract_path: "/tmp/sai-terraform-extract"
    custom_commands:
      download: "wget -O terraform.zip {{url}}"
      extract: "unzip terraform.zip -d {{archive.extract_path}}"
      install: "cp {{archive.extract_path}}/terraform {{install_path}}/terraform"
      uninstall: "rm -f {{install_path}}/terraform"
      validation: "{{install_path}}/terraform version"
      version: "{{install_path}}/terraform version | head -n1"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)
		require.Len(t, saidata.Binaries, 1)

		binary := saidata.Binaries[0]
		assert.Equal(t, "main", binary.Name)
		assert.Equal(t, "https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_amd64.zip", binary.URL)
		assert.Equal(t, "1.6.6", binary.Version)
		assert.Equal(t, "amd64", binary.Architecture)
		assert.Equal(t, "linux", binary.Platform)
		assert.Equal(t, "sha256:b8a3892c58c33ee2b4b8e7c2c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8", binary.Checksum)
		assert.Equal(t, "/usr/local/bin", binary.InstallPath)
		assert.Equal(t, "terraform", binary.Executable)
		assert.Equal(t, "0755", binary.Permissions)

		require.NotNil(t, binary.Archive)
		assert.Equal(t, "zip", binary.Archive.Format)
		assert.Equal(t, "", binary.Archive.StripPrefix)
		assert.Equal(t, "/tmp/sai-terraform-extract", binary.Archive.ExtractPath)

		require.NotNil(t, binary.CustomCommands)
		assert.Equal(t, "wget -O terraform.zip {{url}}", binary.CustomCommands.Download)
		assert.Equal(t, "unzip terraform.zip -d {{archive.extract_path}}", binary.CustomCommands.Extract)
		assert.Equal(t, "cp {{archive.extract_path}}/terraform {{install_path}}/terraform", binary.CustomCommands.Install)
		assert.Equal(t, "rm -f {{install_path}}/terraform", binary.CustomCommands.Uninstall)
		assert.Equal(t, "{{install_path}}/terraform version", binary.CustomCommands.Validation)
		assert.Equal(t, "{{install_path}}/terraform version | head -n1", binary.CustomCommands.Version)
	})

	t.Run("minimal binary configuration", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "simple"
binaries:
  - name: "main"
    url: "https://example.com/binary"
    executable: "simple"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)
		require.Len(t, saidata.Binaries, 1)

		binary := saidata.Binaries[0]
		assert.Equal(t, "main", binary.Name)
		assert.Equal(t, "https://example.com/binary", binary.URL)
		assert.Equal(t, "simple", binary.Executable)
		assert.Empty(t, binary.Version)
		assert.Empty(t, binary.Architecture)
		assert.Empty(t, binary.Platform)
		assert.Nil(t, binary.Archive)
		assert.Nil(t, binary.CustomCommands)
	})
}

func TestScriptType(t *testing.T) {
	t.Run("complete script configuration", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "docker"
scripts:
  - name: "convenience"
    url: "https://get.docker.com"
    version: "24.0.0"
    interpreter: "bash"
    checksum: "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
    arguments: ["--channel", "stable"]
    environment:
      CHANNEL: "stable"
      DOWNLOAD_URL: "https://download.docker.com"
    working_dir: "/tmp"
    timeout: 600
    custom_commands:
      download: "curl -fsSL https://get.docker.com -o get-docker.sh"
      install: "chmod +x get-docker.sh && ./get-docker.sh"
      uninstall: "apt-get remove -y docker-ce"
      validation: "docker --version"
      version: "docker --version | cut -d' ' -f3"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)
		require.Len(t, saidata.Scripts, 1)

		script := saidata.Scripts[0]
		assert.Equal(t, "convenience", script.Name)
		assert.Equal(t, "https://get.docker.com", script.URL)
		assert.Equal(t, "24.0.0", script.Version)
		assert.Equal(t, "bash", script.Interpreter)
		assert.Equal(t, "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7", script.Checksum)
		assert.Len(t, script.Arguments, 2)
		assert.Contains(t, script.Arguments, "--channel")
		assert.Contains(t, script.Arguments, "stable")
		assert.Equal(t, "stable", script.Environment["CHANNEL"])
		assert.Equal(t, "https://download.docker.com", script.Environment["DOWNLOAD_URL"])
		assert.Equal(t, "/tmp", script.WorkingDir)
		assert.Equal(t, 600, script.Timeout)

		require.NotNil(t, script.CustomCommands)
		assert.Equal(t, "curl -fsSL https://get.docker.com -o get-docker.sh", script.CustomCommands.Download)
		assert.Equal(t, "chmod +x get-docker.sh && ./get-docker.sh", script.CustomCommands.Install)
		assert.Equal(t, "apt-get remove -y docker-ce", script.CustomCommands.Uninstall)
		assert.Equal(t, "docker --version", script.CustomCommands.Validation)
		assert.Equal(t, "docker --version | cut -d' ' -f3", script.CustomCommands.Version)
	})

	t.Run("minimal script configuration", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "simple"
scripts:
  - name: "install"
    url: "https://example.com/install.sh"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)
		require.Len(t, saidata.Scripts, 1)

		script := saidata.Scripts[0]
		assert.Equal(t, "install", script.Name)
		assert.Equal(t, "https://example.com/install.sh", script.URL)
		assert.Empty(t, script.Version)
		assert.Empty(t, script.Interpreter)
		assert.Empty(t, script.Checksum)
		assert.Empty(t, script.Arguments)
		assert.Empty(t, script.Environment)
		assert.Empty(t, script.WorkingDir)
		assert.Equal(t, 0, script.Timeout)
		assert.Nil(t, script.CustomCommands)
	})
}

func TestProviderConfigAlternativeProviders(t *testing.T) {
	t.Run("provider config with alternative providers", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "nginx"
providers:
  source:
    sources:
      - name: "main"
        url: "http://nginx.org/download/nginx-1.24.0.tar.gz"
        build_system: "autotools"
  binary:
    binaries:
      - name: "main"
        url: "https://example.com/nginx-binary"
        executable: "nginx"
  script:
    scripts:
      - name: "install"
        url: "https://example.com/install-nginx.sh"
        interpreter: "bash"
`

		saidata, err := LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)

		// Test source provider config
		sourceConfig := saidata.GetProviderConfig("source")
		require.NotNil(t, sourceConfig)
		require.Len(t, sourceConfig.Sources, 1)
		assert.Equal(t, "main", sourceConfig.Sources[0].Name)
		assert.Equal(t, "http://nginx.org/download/nginx-1.24.0.tar.gz", sourceConfig.Sources[0].URL)
		assert.Equal(t, "autotools", sourceConfig.Sources[0].BuildSystem)

		// Test binary provider config
		binaryConfig := saidata.GetProviderConfig("binary")
		require.NotNil(t, binaryConfig)
		require.Len(t, binaryConfig.Binaries, 1)
		assert.Equal(t, "main", binaryConfig.Binaries[0].Name)
		assert.Equal(t, "https://example.com/nginx-binary", binaryConfig.Binaries[0].URL)
		assert.Equal(t, "nginx", binaryConfig.Binaries[0].Executable)

		// Test script provider config
		scriptConfig := saidata.GetProviderConfig("script")
		require.NotNil(t, scriptConfig)
		require.Len(t, scriptConfig.Scripts, 1)
		assert.Equal(t, "install", scriptConfig.Scripts[0].Name)
		assert.Equal(t, "https://example.com/install-nginx.sh", scriptConfig.Scripts[0].URL)
		assert.Equal(t, "bash", scriptConfig.Scripts[0].Interpreter)
	})
}

func TestSoftwareDataGettersAlternativeProviders(t *testing.T) {
	saidata := &SoftwareData{
		Sources: []Source{
			{Name: "src1", URL: "https://example.com/src1.tar.gz", BuildSystem: "autotools"},
			{Name: "src2", URL: "https://example.com/src2.tar.gz", BuildSystem: "cmake"},
		},
		Binaries: []Binary{
			{Name: "bin1", URL: "https://example.com/bin1", Executable: "app1"},
			{Name: "bin2", URL: "https://example.com/bin2", Executable: "app2"},
		},
		Scripts: []Script{
			{Name: "script1", URL: "https://example.com/script1.sh", Interpreter: "bash"},
			{Name: "script2", URL: "https://example.com/script2.sh", Interpreter: "sh"},
		},
	}

	t.Run("GetSourceByName", func(t *testing.T) {
		source := saidata.GetSourceByName("src1")
		require.NotNil(t, source)
		assert.Equal(t, "https://example.com/src1.tar.gz", source.URL)
		assert.Equal(t, "autotools", source.BuildSystem)

		source = saidata.GetSourceByName("nonexistent")
		assert.Nil(t, source)
	})

	t.Run("GetBinaryByName", func(t *testing.T) {
		binary := saidata.GetBinaryByName("bin1")
		require.NotNil(t, binary)
		assert.Equal(t, "https://example.com/bin1", binary.URL)
		assert.Equal(t, "app1", binary.Executable)

		binary = saidata.GetBinaryByName("nonexistent")
		assert.Nil(t, binary)
	})

	t.Run("GetScriptByName", func(t *testing.T) {
		script := saidata.GetScriptByName("script1")
		require.NotNil(t, script)
		assert.Equal(t, "https://example.com/script1.sh", script.URL)
		assert.Equal(t, "bash", script.Interpreter)

		script = saidata.GetScriptByName("nonexistent")
		assert.Nil(t, script)
	})

	t.Run("GetSourceByIndex", func(t *testing.T) {
		source := saidata.GetSourceByIndex(0)
		require.NotNil(t, source)
		assert.Equal(t, "src1", source.Name)

		source = saidata.GetSourceByIndex(1)
		require.NotNil(t, source)
		assert.Equal(t, "src2", source.Name)

		source = saidata.GetSourceByIndex(99)
		assert.Nil(t, source)
	})

	t.Run("GetBinaryByIndex", func(t *testing.T) {
		binary := saidata.GetBinaryByIndex(0)
		require.NotNil(t, binary)
		assert.Equal(t, "bin1", binary.Name)

		binary = saidata.GetBinaryByIndex(1)
		require.NotNil(t, binary)
		assert.Equal(t, "bin2", binary.Name)

		binary = saidata.GetBinaryByIndex(99)
		assert.Nil(t, binary)
	})

	t.Run("GetScriptByIndex", func(t *testing.T) {
		script := saidata.GetScriptByIndex(0)
		require.NotNil(t, script)
		assert.Equal(t, "script1", script.Name)

		script = saidata.GetScriptByIndex(1)
		require.NotNil(t, script)
		assert.Equal(t, "script2", script.Name)

		script = saidata.GetScriptByIndex(99)
		assert.Nil(t, script)
	})
}

func TestAlternativeProvidersJSONSerialization(t *testing.T) {
	saidata := &SoftwareData{
		Version: "0.2",
		Metadata: Metadata{
			Name:        "test",
			Description: "Test software with alternative providers",
		},
		Sources: []Source{
			{
				Name:        "main",
				URL:         "https://example.com/source.tar.gz",
				BuildSystem: "autotools",
				ConfigureArgs: []string{"--enable-ssl"},
				Environment: map[string]string{"CC": "gcc"},
				CustomCommands: &SourceCustomCommands{
					Download:  "wget {{url}}",
					Configure: "./configure {{configure_args | join(' ')}}",
				},
			},
		},
		Binaries: []Binary{
			{
				Name:       "main",
				URL:        "https://example.com/binary.zip",
				Executable: "app",
				Archive: &ArchiveConfig{
					Format:      "zip",
					ExtractPath: "/tmp/extract",
				},
				CustomCommands: &BinaryCustomCommands{
					Download: "curl -L {{url}} -o binary.zip",
					Extract:  "unzip binary.zip",
				},
			},
		},
		Scripts: []Script{
			{
				Name:        "install",
				URL:         "https://example.com/install.sh",
				Interpreter: "bash",
				Arguments:   []string{"--verbose"},
				Environment: map[string]string{"DEBUG": "1"},
				CustomCommands: &ScriptCustomCommands{
					Download: "curl -fsSL {{url}} -o install.sh",
					Install:  "bash install.sh {{arguments | join(' ')}}",
				},
			},
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
	assert.Contains(t, result, "sources")
	assert.Contains(t, result, "binaries")
	assert.Contains(t, result, "scripts")

	// Verify sources structure
	sources := result["sources"].([]interface{})
	require.Len(t, sources, 1)
	source := sources[0].(map[string]interface{})
	assert.Equal(t, "main", source["name"])
	assert.Equal(t, "https://example.com/source.tar.gz", source["url"])
	assert.Equal(t, "autotools", source["build_system"])

	// Verify binaries structure
	binaries := result["binaries"].([]interface{})
	require.Len(t, binaries, 1)
	binary := binaries[0].(map[string]interface{})
	assert.Equal(t, "main", binary["name"])
	assert.Equal(t, "https://example.com/binary.zip", binary["url"])
	assert.Equal(t, "app", binary["executable"])

	// Verify scripts structure
	scripts := result["scripts"].([]interface{})
	require.Len(t, scripts, 1)
	script := scripts[0].(map[string]interface{})
	assert.Equal(t, "install", script["name"])
	assert.Equal(t, "https://example.com/install.sh", script["url"])
	assert.Equal(t, "bash", script["interpreter"])
}

func TestLoadExistingAlternativeProviderSamples(t *testing.T) {
	// Test loading the new alternative provider samples
	saidataFiles := []string{
		"../../docs/saidata_samples/ng/nginx/default.yaml",
		"../../docs/saidata_samples/te/terraform/default.yaml",
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

			// Test alternative provider configurations if present
			if len(saidata.Sources) > 0 {
				for _, source := range saidata.Sources {
					assert.NotEmpty(t, source.Name)
					assert.NotEmpty(t, source.URL)
					if source.BuildSystem != "" {
						assert.Contains(t, []string{"autotools", "cmake", "make", "meson", "ninja", "custom"}, source.BuildSystem)
					}
				}
			}

			if len(saidata.Binaries) > 0 {
				for _, binary := range saidata.Binaries {
					assert.NotEmpty(t, binary.Name)
					assert.NotEmpty(t, binary.URL)
					assert.NotEmpty(t, binary.Executable)
				}
			}

			if len(saidata.Scripts) > 0 {
				for _, script := range saidata.Scripts {
					assert.NotEmpty(t, script.Name)
					assert.NotEmpty(t, script.URL)
				}
			}

			// Validate JSON conversion
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
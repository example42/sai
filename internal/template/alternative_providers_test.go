package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sai-cli/sai/internal/types"
)

func TestSaiSourceTemplateFunction(t *testing.T) {
	saidata := &types.SoftwareData{
		Sources: []types.Source{
			{
				Name:        "main",
				URL:         "http://nginx.org/download/nginx-1.24.0.tar.gz",
				Version:     "1.24.0",
				BuildSystem: "autotools",
				BuildDir:    "/tmp/sai-build-nginx",
				SourceDir:   "/tmp/sai-build-nginx/nginx-1.24.0",
				InstallPrefix: "/usr/local",
				ConfigureArgs: []string{"--with-http_ssl_module", "--with-http_v2_module"},
				BuildArgs:     []string{"-j$(nproc)"},
				InstallArgs:   []string{"install"},
				Prerequisites: []string{"build-essential", "libssl-dev"},
				Environment: map[string]string{
					"CC":     "gcc",
					"CFLAGS": "-O2 -g",
				},
				Checksum: "sha256:5d0b0e8f7e8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f",
				CustomCommands: &types.SourceCustomCommands{
					Download:   "wget -O nginx-1.24.0.tar.gz {{url}}",
					Extract:    "tar -xzf nginx-1.24.0.tar.gz",
					Configure:  "./configure {{configure_args | join(' ')}}",
					Build:      "make {{build_args | join(' ')}}",
					Install:    "make {{install_args | join(' ')}}",
					Uninstall:  "rm -rf /usr/local/sbin/nginx",
					Validation: "nginx -t && nginx -v",
					Version:    "nginx -v 2>&1 | grep -o 'nginx/[0-9.]*'",
				},
			},
			{
				Name:        "alternative",
				URL:         "http://nginx.org/download/nginx-1.25.0.tar.gz",
				Version:     "1.25.0",
				BuildSystem: "cmake",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"source": {
				Sources: []types.Source{
					{
						Name:        "main",
						URL:         "http://nginx.org/download/nginx-1.24.0.tar.gz",
						BuildSystem: "autotools",
						BuildDir:    "/tmp/provider-build",
						Environment: map[string]string{
							"CC": "clang",
						},
					},
				},
			},
		},
	}

	engine := NewTemplateEngine(saidata, "source")

	tests := []struct {
		name     string
		args     []interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "get source name by index",
			args:     []interface{}{0, "name"},
			expected: "main",
			wantErr:  false,
		},
		{
			name:     "get source url by index",
			args:     []interface{}{0, "url"},
			expected: "http://nginx.org/download/nginx-1.24.0.tar.gz",
			wantErr:  false,
		},
		{
			name:     "get source version by index",
			args:     []interface{}{0, "version"},
			expected: "1.24.0",
			wantErr:  false,
		},
		{
			name:     "get source build_system by index",
			args:     []interface{}{0, "build_system"},
			expected: "autotools",
			wantErr:  false,
		},
		{
			name:     "get source build_dir with provider override",
			args:     []interface{}{0, "build_dir"},
			expected: "/tmp/provider-build",
			wantErr:  false,
		},
		{
			name:     "get source environment CC with provider override",
			args:     []interface{}{0, "environment.CC"},
			expected: "clang",
			wantErr:  false,
		},
		{
			name:     "get source environment CFLAGS from default",
			args:     []interface{}{0, "environment.CFLAGS"},
			expected: "-O2 -g",
			wantErr:  false,
		},
		{
			name:     "get configure_args as joined string",
			args:     []interface{}{0, "configure_args"},
			expected: "--with-http_ssl_module --with-http_v2_module",
			wantErr:  false,
		},
		{
			name:     "get build_args as joined string",
			args:     []interface{}{0, "build_args"},
			expected: "-j$(nproc)",
			wantErr:  false,
		},
		{
			name:     "get prerequisites as joined string",
			args:     []interface{}{0, "prerequisites"},
			expected: "build-essential libssl-dev",
			wantErr:  false,
		},
		{
			name:     "get custom command download",
			args:     []interface{}{0, "custom_commands.download"},
			expected: "wget -O nginx-1.24.0.tar.gz {{url}}",
			wantErr:  false,
		},
		{
			name:     "get custom command configure",
			args:     []interface{}{0, "custom_commands.configure"},
			expected: "./configure {{configure_args | join(' ')}}",
			wantErr:  false,
		},
		{
			name:     "get second source by index",
			args:     []interface{}{1, "name"},
			expected: "alternative",
			wantErr:  false,
		},
		{
			name:     "get second source version",
			args:     []interface{}{1, "version"},
			expected: "1.25.0",
			wantErr:  false,
		},
		{
			name:     "get second source build_system",
			args:     []interface{}{1, "build_system"},
			expected: "cmake",
			wantErr:  false,
		},
		{
			name:     "invalid index",
			args:     []interface{}{99, "name"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid field",
			args:     []interface{}{0, "nonexistent_field"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "missing arguments",
			args:     []interface{}{},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid argument types",
			args:     []interface{}{"invalid", "name"},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.saiSource(tt.args...)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaiBinaryTemplateFunction(t *testing.T) {
	saidata := &types.SoftwareData{
		Binaries: []types.Binary{
			{
				Name:         "main",
				URL:          "https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_amd64.zip",
				Version:      "1.6.6",
				Architecture: "amd64",
				Platform:     "linux",
				Checksum:     "sha256:b8a3892c58c33ee2b4b8e7c2c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
				InstallPath:  "/usr/local/bin",
				Executable:   "terraform",
				Permissions:  "0755",
				Archive: &types.ArchiveConfig{
					Format:      "zip",
					StripPrefix: "",
					ExtractPath: "/tmp/sai-terraform-extract",
				},
				CustomCommands: &types.BinaryCustomCommands{
					Download:   "wget -O terraform.zip {{url}}",
					Extract:    "unzip terraform.zip -d {{archive.extract_path}}",
					Install:    "cp {{archive.extract_path}}/terraform {{install_path}}/terraform",
					Uninstall:  "rm -f {{install_path}}/terraform",
					Validation: "{{install_path}}/terraform version",
					Version:    "{{install_path}}/terraform version | head -n1",
				},
			},
			{
				Name:       "alternative",
				URL:        "https://example.com/terraform-alt.tar.gz",
				Version:    "1.7.0",
				Executable: "terraform-alt",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"binary": {
				Binaries: []types.Binary{
					{
						Name:        "main",
						URL:         "https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_darwin_amd64.zip",
						Platform:    "darwin",
						InstallPath: "/usr/local/bin",
					},
				},
			},
		},
	}

	engine := NewTemplateEngine(saidata, "binary")

	tests := []struct {
		name     string
		args     []interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "get binary name by index",
			args:     []interface{}{0, "name"},
			expected: "main",
			wantErr:  false,
		},
		{
			name:     "get binary url with provider override",
			args:     []interface{}{0, "url"},
			expected: "https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_darwin_amd64.zip",
			wantErr:  false,
		},
		{
			name:     "get binary version",
			args:     []interface{}{0, "version"},
			expected: "1.6.6",
			wantErr:  false,
		},
		{
			name:     "get binary architecture",
			args:     []interface{}{0, "architecture"},
			expected: "amd64",
			wantErr:  false,
		},
		{
			name:     "get binary platform with provider override",
			args:     []interface{}{0, "platform"},
			expected: "darwin",
			wantErr:  false,
		},
		{
			name:     "get binary checksum",
			args:     []interface{}{0, "checksum"},
			expected: "sha256:b8a3892c58c33ee2b4b8e7c2c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
			wantErr:  false,
		},
		{
			name:     "get binary install_path",
			args:     []interface{}{0, "install_path"},
			expected: "/usr/local/bin",
			wantErr:  false,
		},
		{
			name:     "get binary executable",
			args:     []interface{}{0, "executable"},
			expected: "terraform",
			wantErr:  false,
		},
		{
			name:     "get binary permissions",
			args:     []interface{}{0, "permissions"},
			expected: "0755",
			wantErr:  false,
		},
		{
			name:     "get archive format",
			args:     []interface{}{0, "archive.format"},
			expected: "zip",
			wantErr:  false,
		},
		{
			name:     "get archive extract_path",
			args:     []interface{}{0, "archive.extract_path"},
			expected: "/tmp/sai-terraform-extract",
			wantErr:  false,
		},
		{
			name:     "get custom command download",
			args:     []interface{}{0, "custom_commands.download"},
			expected: "wget -O terraform.zip {{url}}",
			wantErr:  false,
		},
		{
			name:     "get custom command install",
			args:     []interface{}{0, "custom_commands.install"},
			expected: "cp {{archive.extract_path}}/terraform {{install_path}}/terraform",
			wantErr:  false,
		},
		{
			name:     "get second binary by index",
			args:     []interface{}{1, "name"},
			expected: "alternative",
			wantErr:  false,
		},
		{
			name:     "get second binary version",
			args:     []interface{}{1, "version"},
			expected: "1.7.0",
			wantErr:  false,
		},
		{
			name:     "invalid index",
			args:     []interface{}{99, "name"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid field",
			args:     []interface{}{0, "nonexistent_field"},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.saiBinary(tt.args...)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaiScriptTemplateFunction(t *testing.T) {
	saidata := &types.SoftwareData{
		Scripts: []types.Script{
			{
				Name:        "convenience",
				URL:         "https://get.docker.com",
				Version:     "24.0.0",
				Interpreter: "bash",
				Checksum:    "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7",
				Arguments:   []string{"--channel", "stable"},
				Environment: map[string]string{
					"CHANNEL":      "stable",
					"DOWNLOAD_URL": "https://download.docker.com",
				},
				WorkingDir: "/tmp",
				Timeout:    600,
				CustomCommands: &types.ScriptCustomCommands{
					Download:   "curl -fsSL https://get.docker.com -o get-docker.sh",
					Install:    "chmod +x get-docker.sh && ./get-docker.sh",
					Uninstall:  "apt-get remove -y docker-ce",
					Validation: "docker --version",
					Version:    "docker --version | cut -d' ' -f3",
				},
			},
			{
				Name:        "alternative",
				URL:         "https://example.com/install-alt.sh",
				Version:     "25.0.0",
				Interpreter: "sh",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"script": {
				Scripts: []types.Script{
					{
						Name:       "convenience",
						URL:        "https://get.docker.com",
						WorkingDir: "/tmp/provider-scripts",
						Environment: map[string]string{
							"CHANNEL": "test",
						},
					},
				},
			},
		},
	}

	engine := NewTemplateEngine(saidata, "script")

	tests := []struct {
		name     string
		args     []interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "get script name by index",
			args:     []interface{}{0, "name"},
			expected: "convenience",
			wantErr:  false,
		},
		{
			name:     "get script url",
			args:     []interface{}{0, "url"},
			expected: "https://get.docker.com",
			wantErr:  false,
		},
		{
			name:     "get script version",
			args:     []interface{}{0, "version"},
			expected: "24.0.0",
			wantErr:  false,
		},
		{
			name:     "get script interpreter",
			args:     []interface{}{0, "interpreter"},
			expected: "bash",
			wantErr:  false,
		},
		{
			name:     "get script checksum",
			args:     []interface{}{0, "checksum"},
			expected: "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7",
			wantErr:  false,
		},
		{
			name:     "get script arguments as joined string",
			args:     []interface{}{0, "arguments"},
			expected: "--channel stable",
			wantErr:  false,
		},
		{
			name:     "get script environment CHANNEL with provider override",
			args:     []interface{}{0, "environment.CHANNEL"},
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "get script environment DOWNLOAD_URL from default",
			args:     []interface{}{0, "environment.DOWNLOAD_URL"},
			expected: "https://download.docker.com",
			wantErr:  false,
		},
		{
			name:     "get script working_dir with provider override",
			args:     []interface{}{0, "working_dir"},
			expected: "/tmp/provider-scripts",
			wantErr:  false,
		},
		{
			name:     "get script timeout",
			args:     []interface{}{0, "timeout"},
			expected: "600",
			wantErr:  false,
		},
		{
			name:     "get custom command download",
			args:     []interface{}{0, "custom_commands.download"},
			expected: "curl -fsSL https://get.docker.com -o get-docker.sh",
			wantErr:  false,
		},
		{
			name:     "get custom command install",
			args:     []interface{}{0, "custom_commands.install"},
			expected: "chmod +x get-docker.sh && ./get-docker.sh",
			wantErr:  false,
		},
		{
			name:     "get second script by index",
			args:     []interface{}{1, "name"},
			expected: "alternative",
			wantErr:  false,
		},
		{
			name:     "get second script version",
			args:     []interface{}{1, "version"},
			expected: "25.0.0",
			wantErr:  false,
		},
		{
			name:     "get second script interpreter",
			args:     []interface{}{1, "interpreter"},
			expected: "sh",
			wantErr:  false,
		},
		{
			name:     "invalid index",
			args:     []interface{}{99, "name"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid field",
			args:     []interface{}{0, "nonexistent_field"},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.saiScript(tt.args...)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateFunctionEdgeCases(t *testing.T) {
	t.Run("empty saidata", func(t *testing.T) {
		saidata := &types.SoftwareData{}
		engine := NewTemplateEngine(saidata, "source")

		result, err := engine.saiSource(0, "name")
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("nil custom commands", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name: "main",
					URL:  "https://example.com/source.tar.gz",
					// CustomCommands is nil
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")

		result, err := engine.saiSource(0, "custom_commands.download")
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("nil archive config", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Binaries: []types.Binary{
				{
					Name:       "main",
					URL:        "https://example.com/binary",
					Executable: "app",
					// Archive is nil
				},
			},
		}
		engine := NewTemplateEngine(saidata, "binary")

		result, err := engine.saiBinary(0, "archive.format")
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty environment map", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Scripts: []types.Script{
				{
					Name:        "main",
					URL:         "https://example.com/script.sh",
					Environment: map[string]string{}, // empty map
				},
			},
		}
		engine := NewTemplateEngine(saidata, "script")

		result, err := engine.saiScript(0, "environment.NONEXISTENT")
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

func TestTemplateResolutionOrder(t *testing.T) {
	t.Run("provider override takes precedence", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name:      "main",
					URL:       "https://example.com/default.tar.gz",
					BuildDir:  "/tmp/default-build",
					Environment: map[string]string{
						"CC":     "gcc",
						"CFLAGS": "-O2",
					},
				},
			},
			Providers: map[string]types.ProviderConfig{
				"source": {
					Sources: []types.Source{
						{
							Name:     "main",
							URL:      "https://example.com/provider.tar.gz",
							BuildDir: "/tmp/provider-build",
							Environment: map[string]string{
								"CC": "clang", // Override CC but keep CFLAGS from default
							},
						},
					},
				},
			},
		}

		engine := NewTemplateEngine(saidata, "source")

		// Provider override should take precedence
		url, err := engine.saiSource(0, "url")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/provider.tar.gz", url)

		buildDir, err := engine.saiSource(0, "build_dir")
		require.NoError(t, err)
		assert.Equal(t, "/tmp/provider-build", buildDir)

		cc, err := engine.saiSource(0, "environment.CC")
		require.NoError(t, err)
		assert.Equal(t, "clang", cc)

		// Default value should be used when not overridden
		cflags, err := engine.saiSource(0, "environment.CFLAGS")
		require.NoError(t, err)
		assert.Equal(t, "-O2", cflags)
	})
}
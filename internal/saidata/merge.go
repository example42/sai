package saidata

import "sai/internal/types"

// mergePackages merges package slices, replacing packages with same name
func mergePackages(base, override []types.Package) []types.Package {
	result := make([]types.Package, 0, len(base)+len(override))
	
	// Create a map for quick lookup
	overrideMap := make(map[string]types.Package)
	for _, pkg := range override {
		overrideMap[pkg.Name] = pkg
	}
	
	// Add base packages, replacing with override if exists
	for _, pkg := range base {
		if overridePkg, exists := overrideMap[pkg.Name]; exists {
			result = append(result, overridePkg)
			delete(overrideMap, pkg.Name) // Mark as processed
		} else {
			result = append(result, pkg)
		}
	}
	
	// Add remaining override packages
	for _, pkg := range overrideMap {
		result = append(result, pkg)
	}
	
	return result
}

// mergeServices merges service slices, replacing services with same name
func mergeServices(base, override []types.Service) []types.Service {
	result := make([]types.Service, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.Service)
	for _, svc := range override {
		overrideMap[svc.Name] = svc
	}
	
	for _, svc := range base {
		if overrideSvc, exists := overrideMap[svc.Name]; exists {
			result = append(result, overrideSvc)
			delete(overrideMap, svc.Name)
		} else {
			result = append(result, svc)
		}
	}
	
	for _, svc := range overrideMap {
		result = append(result, svc)
	}
	
	return result
}

// mergeFiles merges file slices, replacing files with same name
func mergeFiles(base, override []types.File) []types.File {
	result := make([]types.File, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.File)
	for _, file := range override {
		overrideMap[file.Name] = file
	}
	
	for _, file := range base {
		if overrideFile, exists := overrideMap[file.Name]; exists {
			result = append(result, overrideFile)
			delete(overrideMap, file.Name)
		} else {
			result = append(result, file)
		}
	}
	
	for _, file := range overrideMap {
		result = append(result, file)
	}
	
	return result
}

// mergeDirectories merges directory slices, replacing directories with same name
func mergeDirectories(base, override []types.Directory) []types.Directory {
	result := make([]types.Directory, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.Directory)
	for _, dir := range override {
		overrideMap[dir.Name] = dir
	}
	
	for _, dir := range base {
		if overrideDir, exists := overrideMap[dir.Name]; exists {
			result = append(result, overrideDir)
			delete(overrideMap, dir.Name)
		} else {
			result = append(result, dir)
		}
	}
	
	for _, dir := range overrideMap {
		result = append(result, dir)
	}
	
	return result
}

// mergeCommands merges command slices, replacing commands with same name
func mergeCommands(base, override []types.Command) []types.Command {
	result := make([]types.Command, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.Command)
	for _, cmd := range override {
		overrideMap[cmd.Name] = cmd
	}
	
	for _, cmd := range base {
		if overrideCmd, exists := overrideMap[cmd.Name]; exists {
			result = append(result, overrideCmd)
			delete(overrideMap, cmd.Name)
		} else {
			result = append(result, cmd)
		}
	}
	
	for _, cmd := range overrideMap {
		result = append(result, cmd)
	}
	
	return result
}

// mergePorts merges port slices, replacing ports with same port number
func mergePorts(base, override []types.Port) []types.Port {
	result := make([]types.Port, 0, len(base)+len(override))
	
	overrideMap := make(map[int]types.Port)
	for _, port := range override {
		overrideMap[port.Port] = port
	}
	
	for _, port := range base {
		if overridePort, exists := overrideMap[port.Port]; exists {
			result = append(result, overridePort)
			delete(overrideMap, port.Port)
		} else {
			result = append(result, port)
		}
	}
	
	for _, port := range overrideMap {
		result = append(result, port)
	}
	
	return result
}

// mergeContainers merges container slices, replacing containers with same name
func mergeContainers(base, override []types.Container) []types.Container {
	result := make([]types.Container, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.Container)
	for _, container := range override {
		overrideMap[container.Name] = container
	}
	
	for _, container := range base {
		if overrideContainer, exists := overrideMap[container.Name]; exists {
			result = append(result, overrideContainer)
			delete(overrideMap, container.Name)
		} else {
			result = append(result, container)
		}
	}
	
	for _, container := range overrideMap {
		result = append(result, container)
	}
	
	return result
}

// mergeProviderConfig merges provider configurations
func mergeProviderConfig(base, override types.ProviderConfig) types.ProviderConfig {
	result := base
	
	// Merge simple slices by replacement
	if len(override.Prerequisites) > 0 {
		result.Prerequisites = override.Prerequisites
	}
	if len(override.BuildCommands) > 0 {
		result.BuildCommands = override.BuildCommands
	}
	
	// Merge complex slices using the same logic as main saidata
	if len(override.Packages) > 0 {
		result.Packages = mergePackages(result.Packages, override.Packages)
	}
	if len(override.Services) > 0 {
		result.Services = mergeServices(result.Services, override.Services)
	}
	if len(override.Files) > 0 {
		result.Files = mergeFiles(result.Files, override.Files)
	}
	if len(override.Directories) > 0 {
		result.Directories = mergeDirectories(result.Directories, override.Directories)
	}
	if len(override.Commands) > 0 {
		result.Commands = mergeCommands(result.Commands, override.Commands)
	}
	if len(override.Ports) > 0 {
		result.Ports = mergePorts(result.Ports, override.Ports)
	}
	if len(override.Containers) > 0 {
		result.Containers = mergeContainers(result.Containers, override.Containers)
	}
	
	// Merge package sources and repositories
	if len(override.PackageSources) > 0 {
		result.PackageSources = mergePackageSources(result.PackageSources, override.PackageSources)
	}
	if len(override.Repositories) > 0 {
		result.Repositories = mergeRepositories(result.Repositories, override.Repositories)
	}
	
	return result
}

// mergePackageSources merges package source slices, replacing sources with same name
func mergePackageSources(base, override []types.PackageSource) []types.PackageSource {
	result := make([]types.PackageSource, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.PackageSource)
	for _, src := range override {
		overrideMap[src.Name] = src
	}
	
	for _, src := range base {
		if overrideSrc, exists := overrideMap[src.Name]; exists {
			result = append(result, overrideSrc)
			delete(overrideMap, src.Name)
		} else {
			result = append(result, src)
		}
	}
	
	for _, src := range overrideMap {
		result = append(result, src)
	}
	
	return result
}

// mergeRepositories merges repository slices, replacing repositories with same name
func mergeRepositories(base, override []types.Repository) []types.Repository {
	result := make([]types.Repository, 0, len(base)+len(override))
	
	overrideMap := make(map[string]types.Repository)
	for _, repo := range override {
		overrideMap[repo.Name] = repo
	}
	
	for _, repo := range base {
		if overrideRepo, exists := overrideMap[repo.Name]; exists {
			result = append(result, overrideRepo)
			delete(overrideMap, repo.Name)
		} else {
			result = append(result, repo)
		}
	}
	
	for _, repo := range overrideMap {
		result = append(result, repo)
	}
	
	return result
}

// mergeCompatibility merges compatibility information
func mergeCompatibility(base, override *types.Compatibility) *types.Compatibility {
	result := *base
	
	// Override matrix completely if provided
	if len(override.Matrix) > 0 {
		result.Matrix = override.Matrix
	}
	
	// Merge version information
	if override.Versions != nil {
		if result.Versions == nil {
			result.Versions = override.Versions
		} else {
			if override.Versions.Latest != "" {
				result.Versions.Latest = override.Versions.Latest
			}
			if override.Versions.Minimum != "" {
				result.Versions.Minimum = override.Versions.Minimum
			}
			if override.Versions.LatestLTS != "" {
				result.Versions.LatestLTS = override.Versions.LatestLTS
			}
			if override.Versions.LatestMinimum != "" {
				result.Versions.LatestMinimum = override.Versions.LatestMinimum
			}
		}
	}
	
	return &result
}
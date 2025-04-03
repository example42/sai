package pkgmanager

// GetProvider creates the appropriate package manager provider based on the name
func GetProvider(name string) Provider {
	switch name {
	case "rpm":
		return NewRPMProvider()
	case "apt":
		return NewAPTProvider()
	case "brew":
		return NewBrewProvider()
	case "winget":
		return NewWingetProvider()
	case "pacman":
		return NewPacmanProvider()
	case "zypper":
		return NewZypperProvider()
	default:
		// Return APT provider as default
		return NewAPTProvider()
	}
}

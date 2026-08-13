// Package version hält die zentrale Projekt- und Modulversion.
// Quell der Wahrheit ist die VERSION.json im Repo-Root; diese Datei wird
// von scripts/bump-version.py synchronisiert und von check-version.py geprüft.
package version

const (
	Project = "0.2.0"

	Bus          = "0.2.0-bus.1"
	SCAutomation = "0.2.0-s-c.1"
	SEMonitor    = "0.2.0-s-e.1"
)

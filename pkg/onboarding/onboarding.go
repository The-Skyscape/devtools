// Package onboarding provides helpers for creating user onboarding experiences
// with tour modals and first-time user flows.
package onboarding

import (
	"github.com/The-Skyscape/devtools/pkg/settings"
)

// TourManager helps manage tour/onboarding state for users
type TourManager struct {
	settings *settings.Manager
	prefix   string // Prefix for tour-related settings
}

// NewTourManager creates a new tour manager
func NewTourManager(settingsManager *settings.Manager, tourName string) *TourManager {
	return &TourManager{
		settings: settingsManager,
		prefix:   tourName + "_",
	}
}

// ShouldShowTour checks if the tour should be displayed to the user
func (tm *TourManager) ShouldShowTour() bool {
	// Check if user explicitly said never show again
	if tm.settings.GetBool(tm.prefix + "never_show") {
		return false
	}
	
	// Check if tour was completed
	return !tm.settings.GetBool(tm.prefix + "completed")
}

// CompleteTour marks the tour as completed
func (tm *TourManager) CompleteTour(neverShowAgain bool) error {
	// Save completion status
	if err := tm.settings.SetBool(tm.prefix+"completed", true, "user_preference"); err != nil {
		return err
	}
	
	// Save never show preference if requested
	if neverShowAgain {
		if err := tm.settings.SetBool(tm.prefix+"never_show", true, "user_preference"); err != nil {
			return err
		}
	}
	
	return nil
}

// ResetTour resets the tour state (useful for testing or re-onboarding)
func (tm *TourManager) ResetTour() error {
	if err := tm.settings.Delete(tm.prefix + "completed"); err != nil {
		return err
	}
	return tm.settings.Delete(tm.prefix + "never_show")
}

// GetTourStep gets the current step in a multi-step tour
func (tm *TourManager) GetTourStep() int {
	step, _ := tm.settings.Get(tm.prefix + "current_step")
	if step == "" {
		return 1
	}
	// Convert string to int (simplified for example)
	switch step {
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	default:
		return 1
	}
}

// SetTourStep saves the current step in a multi-step tour
func (tm *TourManager) SetTourStep(step int) error {
	stepStr := "1"
	switch step {
	case 2:
		stepStr = "2"
	case 3:
		stepStr = "3"
	case 4:
		stepStr = "4"
	}
	return tm.settings.Set(tm.prefix+"current_step", stepStr, "user_preference")
}
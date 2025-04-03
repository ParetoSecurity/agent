package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
)

func createChecksTab() *container.TabItem {
	checksContent := container.NewVBox()
	for _, claim := range claims.All {
		checksContent.Add(widget.NewLabel(claim.Title))
		for _, check := range claim.Checks {
			chkbox := widget.NewCheck(check.Name(), func(state bool) {
				if state {
					shared.EnableCheck(check.UUID())
				} else {
					shared.DisableCheck(check.UUID())
				}
			})
			chkbox.SetChecked(!shared.IsCheckDisabled(check.UUID()))
			checksContent.Add(chkbox)
		}
	}
	return container.NewTabItem("Checks", container.NewPadded(container.NewScroll(checksContent)))
}

func createAboutTab() *container.TabItem {
	aboutContent := container.NewCenter(container.NewVBox(
		canvas.NewText("ParetoSecurity Agent", theme.Color(theme.ColorNameForeground)),
		canvas.NewText(fmt.Sprintf("Commit: %s", shared.Commit), theme.Color(theme.ColorNameForeground)),
		canvas.NewText(fmt.Sprintf("Version: %s", shared.Version), theme.Color(theme.ColorNameForeground)),
		canvas.NewText("Made with ❤️ at Niteo", theme.Color(theme.ColorNameForeground)),
	))
	resource := fyne.NewStaticResource("pareto", shared.IconLarge)
	image := canvas.NewImageFromResource(resource)
	image.SetMinSize(fyne.NewSize(160, 160))
	image.FillMode = canvas.ImageFillContain
	content := container.NewHBox(image, aboutContent)
	return container.NewTabItem("About", container.NewCenter(content))
}

func createGeneralTab() *container.TabItem {
	generalContent := container.NewVBox()

	// Automatically launch on system startup
	startupToggle := widget.NewCheck("Automatically launch on system startup", func(state bool) {
		// Handle enabling/disabling startup launch
	})
	startupToggle.SetChecked(true) // Default to enabled
	generalContent.Add(container.NewVBox(
		startupToggle,
		widget.NewLabel("To enable continuous monitoring and reporting."), // Footer description
	))

	// Use alternative color scheme
	colorSchemeToggle := widget.NewCheck("Use alternative color scheme", func(state bool) {
		// Handle enabling/disabling alternative color scheme
	})
	colorSchemeToggle.SetChecked(false) // Default to disabled
	generalContent.Add(container.NewVBox(
		colorSchemeToggle,
		widget.NewLabel("Improve default colors for accessibility."), // Footer description
	))

	// Run checks in the background
	backgroundChecksToggle := widget.NewCheck("Run checks in the background", func(state bool) {
		// Handle enabling/disabling background checks
	})
	backgroundChecksToggle.SetChecked(true) // Default to enabled
	generalContent.Add(container.NewVBox(
		backgroundChecksToggle,
		widget.NewLabel("Enable continuous checks without user interaction."), // Footer description
	))

	return container.NewTabItem("General", container.NewPadded(container.NewScroll(generalContent)))
}

func generateTabs(window fyne.Window) *container.AppTabs {
	// Create tabs
	generalTab := createGeneralTab() // New tab
	checksTab := createChecksTab()
	aboutTab := createAboutTab()

	// Create a tab container
	tabs := container.NewAppTabs(generalTab, checksTab, aboutTab)
	tabs.SetTabLocation(container.TabLocationTop)

	// Adjust window size based on selected tab
	tabs.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "Checks":
			window.Resize(fyne.NewSize(420, 450))
		case "About":
			window.Resize(fyne.NewSize(420, 200))
		case "General":
			window.Resize(fyne.NewSize(420, 210))
		}
	}

	return tabs
}

func CreatePreferencesWindow() {
	// Create a new Fyne application
	application := app.New()
	window := application.NewWindow("Preferences")

	// Disable minimize by fixing the window size
	window.SetFixedSize(true)
	window.SetMaster()

	// Set the tab container as the window content
	window.SetContent(generateTabs(window))

	// Show the window and run the application
	window.Resize(fyne.NewSize(420, 210)) // Default to Checks tab size
	window.ShowAndRun()
}

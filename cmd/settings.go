package cmd

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ParetoSecurity/agent/claims"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/spf13/cobra"
)

var preferencesUICmd = &cobra.Command{
	Use:   "preferences",
	Short: "Display the preferences dialog",
	Run: func(cc *cobra.Command, args []string) {
		// Create a new Fyne application
		application := app.New()
		window := application.NewWindow("Preferences")

		// Disable minimize by fixing the window size
		window.SetFixedSize(true)

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
				window.Resize(fyne.NewSize(420, 150))
			case "General":
				window.Resize(fyne.NewSize(420, 450))
			}
		}

		// Set the tab container as the window content
		window.SetContent(tabs)

		// Show the window and run the application
		window.Resize(fyne.NewSize(420, 450)) // Default to Checks tab size
		window.ShowAndRun()
	},
}

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
	aboutContent := container.NewVBox(
		widget.NewLabel("ParetoSecurity Agent"),
		widget.NewLabel(fmt.Sprintf("Chanel: %s", shared.Config.TeamID)),
		widget.NewLabel(fmt.Sprintf("Commit: %s", shared.Commit)),
		widget.NewLabel(fmt.Sprintf("Version: %s", shared.Version)),
		widget.NewLabel("Made with (heart) at Niteo"),
	)
	resource := fyne.NewStaticResource("pareto", shared.IconLarge)
	image := canvas.NewImageFromResource(resource)
	image.SetMinSize(fyne.NewSize(128, 128))
	image.FillMode = canvas.ImageFillContain
	content := container.NewHBox(image, aboutContent)
	return container.NewTabItem("About", container.NewPadded(container.NewCenter(content)))
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

	// Only show in menu bar when the checks are failing
	menuBarToggle := widget.NewCheck("Only show in menu bar when the checks are failing", func(state bool) {
		// Handle enabling/disabling menu bar visibility
	})
	menuBarToggle.SetChecked(false) // Default to disabled
	generalContent.Add(container.NewVBox(
		menuBarToggle,
		widget.NewLabel("To show the menu bar icon, launch the app again."), // Footer description
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

	return container.NewTabItem("General", container.NewPadded(container.NewScroll(generalContent)))
}

func init() {
	rootCmd.AddCommand(preferencesUICmd)
}

# Pareto Security Tray Application

This is the Windows system tray application for Pareto Security that also handles protocol URLs for device linking.

## Prerequisites

1. Install Wails 3:
```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.29
```

2. Install Node.js dependencies:
```bash
cd ui
yarn install
```

## Building the Application

### 1. Generate Wails Bindings

From the `cmd/paretosecurity-tray` directory, generate the TypeScript bindings for the Go services:

```bash
wails3 generate bindings -v -ts -d ./ui/src/.binding -names -b
```

This will create the necessary TypeScript/JavaScript files that allow the Elm UI to communicate with the Go backend services.

### 2. Generate Wails Runtime

Generate the Wails runtime files:

```bash
wails3 generate runtime -d ./ui/wails
```

This creates the runtime JavaScript files needed for the Wails framework to function properly.

### 3. Build the UI

Build the Elm UI and create the distribution files:

```bash
cd ui
yarn build
```

This will:
- Compile the Elm code to JavaScript
- Bundle all assets with Vite
- Create the `dist` folder with all necessary files
- The `dist` folder will be embedded into the Go executable

### 4. Build the Go Executable

From the `cmd/paretosecurity-tray` directory:

```bash
go build
```

This will create `paretosecurity-tray.exe` with the embedded UI assets.

## Features

### System Tray Mode
When run without arguments, the application starts as a system tray icon that:
- Shows security check status
- Allows manual security checks
- Provides options for startup and team linking
- Shows help and documentation links

### Link Command Mode
When run with the `link` command and a URL, it opens a GUI window for device linking:

```bash
paretosecurity-tray.exe link "paretosecurity://linkDevice?invite_id=XXX"
```

This is used by the Windows protocol handler to process `paretosecurity://` URLs from web browsers.

## Protocol Handler

The installer registers `paretosecurity://` URLs to be handled by:
```
paretosecurity-tray.exe link "%1"
```

This allows users to click enrollment links in their browser and have them automatically open the linking UI.

## Development

### UI Structure
- `ui/src/Link.elm` - Elm application for the linking UI
- `ui/src/Link.css` - TailwindCSS/DaisyUI styles
- `ui/index.html` - HTML entry point
- `linkservice.go` - Go service that handles device linking
- `linkapp.go` - Wails application configuration for link mode
- `main.go` - Main entry point that routes between tray and link modes

### Modifying the UI

1. Edit the Elm files in `ui/src/`
2. Rebuild with `yarn build` in the `ui` directory
3. Rebuild the Go executable to embed the new UI

### Testing

To test the link command locally:
```bash
go run . link "paretosecurity://linkDevice?invite_id=test123"
```
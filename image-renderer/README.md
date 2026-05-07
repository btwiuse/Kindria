# Bubble Tea v2 Image Renderer Module

This directory is a standalone Go module for the reusable terminal image rendering component originally extracted from Kindria's TUI.

Module path:

```text
github.com/btwiuse/Kindria/image-renderer
```

## Included Content

- reusable package: `./renderer.go`
- standalone example: `./examples/minimal`
- module-scoped dependencies upgraded to **Bubble Tea v2**

## Dependencies

This module is intentionally isolated from the root application module and uses Bubble Tea v2-era imports:

- `charm.land/bubbletea/v2`
- `charm.land/lipgloss/v2`
- `github.com/blacktop/go-termimg`
- `github.com/disintegration/imaging`

## Implementation Principles

### 1) Async render scheduling

`(*ImageRenderer).Sync(SyncRequest)` converts visible image tasks into a `tea.Batch(...)` of asynchronous render commands.

Each command:

1. opens the source image;
2. computes target pixel size;
3. resizes and pads the image if needed;
4. renders the image through `go-termimg`;
5. returns an internal Bubble Tea message that is consumed by `Update(msg)`.

### 2) Terminal-adaptive sizing

The component renders by character-cell size (`Width` / `Height`) but resolves a pixel target using this priority:

1. explicit `CellPixelWidth` / `CellPixelHeight`;
2. `termimg.QueryTerminalFeatures()` font metrics;
3. fallback to char-cell dimensions directly.

This makes the same API work both in fully known layouts and in terminals where only runtime feature probing is available.

### 3) Cache and in-flight deduplication

The renderer maintains three maps:

- `cache`: rendered output keyed by source identity + size + protocol
- `pending`: in-flight renders so the same task is not scheduled repeatedly
- `rendered`: latest rendered output by logical task ID

`RenderTask.CacheKey` lets callers define stable logical identity across UI refreshes.

### 4) Overlay composition

`Overlay([]OverlayPlacement)` produces ANSI cursor-positioned output (`\x1b[row;colH...`) for already-rendered images.

Typical Bubble Tea flow:

1. render your base layout as text;
2. compute row/column positions for images;
3. append the overlay string to the base content;
4. return the final content in `tea.View`.

## Public API

```go
renderer := imagerenderer.New()

cmd := renderer.Sync(imagerenderer.SyncRequest{...})
handled := renderer.Update(msg)
raw := renderer.Rendered("cover-1")
overlay := renderer.Overlay([]imagerenderer.OverlayPlacement{...})
renderer.ResetVisible()
```

## Bubble Tea v2 Usage Notes

This module and its example follow the Bubble Tea v2 migration guide:

- import `tea` from `charm.land/bubbletea/v2`
- import Lip Gloss from `charm.land/lipgloss/v2`
- use `tea.KeyPressMsg` instead of `tea.KeyMsg`
- return `tea.View` instead of `string`
- declare alternate screen in `View()` rather than `tea.WithAltScreen()`

## Run the Example

```bash
cd /home/runner/work/Kindria/Kindria/image-renderer
go run ./examples/minimal
```

## Validate the Module

```bash
cd /home/runner/work/Kindria/Kindria/image-renderer
go test ./...
go build ./examples/minimal
```

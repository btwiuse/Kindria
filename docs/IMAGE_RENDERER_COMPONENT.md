# ImageRenderer Component

`internal/tui/components/image_renderer.go` provides a reusable Bubble Tea image rendering component for terminal UIs.

## Purpose

The component extracts image-specific logic from screen models, so your Bubble Tea model only needs to:

1. request async image rendering tasks;
2. consume render-complete messages;
3. place rendered image output with row/column coordinates.

This keeps rendering concerns isolated from business/UI state.

## Implementation Principles

## 1) Async rendering via Bubble Tea commands

- `Sync(SyncRequest)` builds a `tea.Batch(...)` of render commands.
- Each command opens image source (`imaging.Open`), resizes it, renders via `go-termimg`, and returns internal message `coverLoadedMsg`.
- Bubble Tea event loop receives those messages; your model forwards them to `ImageRenderer.Update(msg)`.

## 2) Terminal-adaptive sizing

`SyncRequest` accepts character-cell size (`Width`/`Height`) and optional cell-pixel size (`CellPixelWidth`/`CellPixelHeight`).

Sizing strategy:

- if cell-pixel size is known, use `Width*CellPixelWidth` and `Height*CellPixelHeight`;
- else fallback to `termimg.QueryTerminalFeatures()` font metrics;
- else fallback to char-cell size directly.

## 3) Cache + in-flight deduplication

- `cache` stores rendered outputs by a cache key containing source identity + render dimensions + protocol.
- `pending` prevents duplicate concurrent renders for the same key.
- `rendered` stores current visible outputs by logical task `ID`.

`RenderTask.CacheKey` lets caller define logical identity (for example, include book identity + path), avoiding accidental cache collisions.

## 4) Overlay drawing

- `Overlay([]OverlayPlacement)` returns ANSI cursor-positioned output (`\x1b[row;colH...`).
- Caller appends this overlay string to base layout string in `View()`.
- `ID` in `OverlayPlacement` maps to data in `rendered`.

## API Quick Reference

- `NewImageRenderer() *ImageRenderer`
- `(*ImageRenderer).Sync(SyncRequest) tea.Cmd`
- `(*ImageRenderer).Update(msg tea.Msg) bool`
- `(*ImageRenderer).Rendered(id string) string`
- `(*ImageRenderer).Overlay([]OverlayPlacement) string`
- `(*ImageRenderer).ResetVisible()`

## Usage Flow (Bubble Tea)

1. Initialize renderer in your model.
2. Trigger `Sync(...)` when visible image set / size changes.
3. In `Update`, call `renderer.Update(msg)` first.
4. In `View`, render base cards/panels first, then append `renderer.Overlay(...)`.
5. Call `ResetVisible()` when switching to a new dataset/view.

## Minimal Runnable Example

See:

- `examples/image_renderer_minimal/main.go`

Run:

```bash
go run ./examples/image_renderer_minimal
```

Press `q` / `ctrl+c` to quit.

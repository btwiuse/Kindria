package components

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/blacktop/go-termimg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/disintegration/imaging"
)

type RenderTask struct {
	ID         string
	SourcePath string
}

type SyncRequest struct {
	Tasks           []RenderTask
	Width           int
	Height          int
	CellPixelWidth  int
	CellPixelHeight int
}

type OverlayPlacement struct {
	ID  string
	Row int
	Col int
}

type coverLoadedMsg struct {
	id   string
	key  string
	data string
}

type ImageRenderer struct {
	rendered map[string]string
	cache    map[string]string
	pending  map[string]struct{}
}

func NewImageRenderer() *ImageRenderer {
	return &ImageRenderer{
		rendered: make(map[string]string),
		cache:    make(map[string]string),
		pending:  make(map[string]struct{}),
	}
}

func (r *ImageRenderer) ResetVisible() {
	r.rendered = make(map[string]string)
	r.pending = make(map[string]struct{})
}

func (r *ImageRenderer) Rendered(id string) string {
	return r.rendered[id]
}

func (r *ImageRenderer) Update(msg tea.Msg) bool {
	coverMsg, ok := msg.(coverLoadedMsg)
	if !ok {
		return false
	}

	delete(r.pending, coverMsg.key)
	r.cache[coverMsg.key] = coverMsg.data
	r.rendered[coverMsg.id] = coverMsg.data
	return true
}

func (r *ImageRenderer) Sync(req SyncRequest) tea.Cmd {
	if len(req.Tasks) == 0 || req.Width <= 0 || req.Height <= 0 {
		return nil
	}

	protocol := termimg.DetectProtocol()
	features := termimg.QueryTerminalFeatures()
	targetPixelWidth := req.Width
	targetPixelHeight := req.Height
	if req.CellPixelWidth > 0 && req.CellPixelHeight > 0 {
		targetPixelWidth = req.Width * req.CellPixelWidth
		targetPixelHeight = req.Height * req.CellPixelHeight
	} else if features != nil && features.FontWidth > 0 && features.FontHeight > 0 {
		targetPixelWidth = req.Width * features.FontWidth
		targetPixelHeight = req.Height * features.FontHeight
	}
	if targetPixelWidth <= 0 {
		targetPixelWidth = req.Width
	}
	if targetPixelHeight <= 0 {
		targetPixelHeight = req.Height
	}

	cmds := make([]tea.Cmd, 0, len(req.Tasks))
	for _, task := range req.Tasks {
		cacheKey := fmt.Sprintf("%s|%dx%d|%dx%d|%v", task.SourcePath, req.Width, req.Height, targetPixelWidth, targetPixelHeight, protocol)
		if cached, ok := r.cache[cacheKey]; ok {
			r.rendered[task.ID] = cached
			continue
		}
		if _, pending := r.pending[cacheKey]; pending {
			continue
		}
		r.pending[cacheKey] = struct{}{}

		taskID := task.ID
		coverPath := task.SourcePath
		key := cacheKey
		cmds = append(cmds, func() tea.Msg {
			srcImage, err := imaging.Open(coverPath)
			if err != nil {
				return coverLoadedMsg{id: taskID, key: key, data: ""}
			}

			resizedImage := imaging.Fit(srcImage, targetPixelWidth, targetPixelHeight, imaging.Lanczos)
			if resizedImage.Bounds().Dx() != targetPixelWidth || resizedImage.Bounds().Dy() != targetPixelHeight {
				canvas := imaging.New(targetPixelWidth, targetPixelHeight, color.NRGBA{R: 10, G: 10, B: 10, A: 255})
				resizedImage = imaging.PasteCenter(canvas, resizedImage)
			}

			img := termimg.New(resizedImage).Scale(termimg.ScaleNone)
			if protocol == termimg.Halfblocks {
				img = img.Dither(true).DitherMode(termimg.DitherFloydSteinberg)
			}

			cover := termimg.NewImageWidget(img)
			if cover == nil {
				return coverLoadedMsg{id: taskID, key: key, data: ""}
			}
			cover.SetSize(req.Width, req.Height).SetProtocol(protocol)
			coverRendered, err := cover.Render()
			if err != nil {
				log.Printf("Err rendering cover: %v ", err)
				return coverLoadedMsg{id: taskID, key: key, data: ""}
			}
			return coverLoadedMsg{id: taskID, key: key, data: coverRendered}
		})
	}

	return tea.Batch(cmds...)
}

func (r *ImageRenderer) Overlay(placements []OverlayPlacement) string {
	if len(placements) == 0 {
		return ""
	}
	var overlay strings.Builder
	for _, placement := range placements {
		rendered := r.rendered[placement.ID]
		if rendered == "" {
			continue
		}
		overlay.WriteString("\x1b[")
		overlay.WriteString(strconv.Itoa(placement.Row))
		overlay.WriteString(";")
		overlay.WriteString(strconv.Itoa(placement.Col))
		overlay.WriteString("H")
		overlay.WriteString(rendered)
	}
	return overlay.String()
}

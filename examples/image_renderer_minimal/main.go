package main

import (
	"Kindria/internal/tui/components"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const demoTaskID = "demo-image"

type model struct {
	renderer *components.ImageRenderer
	image    string
}

func newModel() (*model, error) {
	imgPath, err := createDemoImage()
	if err != nil {
		return nil, err
	}
	return &model{
		renderer: components.NewImageRenderer(),
		image:    imgPath,
	}, nil
}

func (m *model) Init() tea.Cmd {
	return m.renderer.Sync(components.SyncRequest{
		Tasks: []components.RenderTask{
			{
				ID:         demoTaskID,
				CacheKey:   "minimal-demo|" + m.image,
				SourcePath: m.image,
			},
		},
		Width:  28,
		Height: 14,
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.renderer.Update(msg) {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			_ = os.Remove(m.image)
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) View() string {
	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(40).
		Height(18).
		Render("")

	base := "ImageRenderer minimal example (press q to quit)\n\n" + frame
	overlay := m.renderer.Overlay([]components.OverlayPlacement{
		{ID: demoTaskID, Row: 4, Col: 3},
	})
	return base + overlay
}

func createDemoImage() (string, error) {
	img := image.NewNRGBA(image.Rect(0, 0, 640, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			r := uint8((x * 255) / 639)
			g := uint8((y * 255) / 359)
			b := uint8(180)
			img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	f, err := os.CreateTemp("", "kindria-image-renderer-*.png")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func main() {
	m, err := newModel()
	if err != nil {
		fmt.Println("failed to initialize demo image:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("program failed:", err)
		os.Exit(1)
	}
}

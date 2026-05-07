package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	imagerenderer "github.com/btwiuse/Kindria/image-renderer"
)

const demoTaskID = "demo-image"

type model struct {
	renderer *imagerenderer.ImageRenderer
	image    string
}

func newModel() (*model, error) {
	imgPath, err := createDemoImage()
	if err != nil {
		return nil, err
	}

	return &model{
		renderer: imagerenderer.New(),
		image:    imgPath,
	}, nil
}

func (m *model) Init() tea.Cmd {
	return m.renderer.Sync(imagerenderer.SyncRequest{
		Tasks: []imagerenderer.RenderTask{
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
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			_ = os.Remove(m.image)
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) View() tea.View {
	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(40).
		Height(18).
		Render("")

	base := "ImageRenderer minimal example (Bubble Tea v2, press q to quit)\n\n" + frame
	overlay := m.renderer.Overlay([]imagerenderer.OverlayPlacement{
		{ID: demoTaskID, Row: 4, Col: 3},
	})

	view := tea.NewView(base + overlay)
	view.AltScreen = true
	return view
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

	f, err := os.CreateTemp("", "kindria-image-renderer-v2-*.png")
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

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("program failed:", err)
		os.Exit(1)
	}
}

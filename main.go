package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	checkStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).SetString("🗸")
	crossStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).SetString("✘")
	spinnerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	glowStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	downloadedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("157"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	marginStyle     = lipgloss.NewStyle().Margin(1, 3)
	downloadPath    = "mods/%s"
)

type modDownloaded string

type model struct {
	width    int
	height   int
	index    int
	err      bool
	done     bool
	spinner  spinner.Model
	progress progress.Model
	manifest Manifest
	errors   []string
}

func newModel(manifest Manifest) model {
	s := spinner.New()
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s.Style = spinnerStyle
	return model{spinner: s, progress: p, manifest: manifest}
}

func (m model) Init() tea.Cmd {
	if _, err := os.Stat(fmt.Sprintf(downloadPath, "")); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(fmt.Sprintf(downloadPath, ""), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	return tea.Batch(
		tea.Printf("Minecraft version: %s\nLoader: %s\nModpack: %s Version: %s\nAuthor: %s\n",
			glowStyle.Render(m.manifest.Minecraft.Version),
			glowStyle.Render(m.manifest.Minecraft.ModLoaders[0].ID),
			glowStyle.Render(m.manifest.Name),
			glowStyle.Render(m.manifest.Version),
			glowStyle.Render(m.manifest.Author),
		),
		tea.Tick(1*time.Millisecond, func(t time.Time) tea.Msg { return modDownloaded("") }),
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case modDownloaded:
		if m.index >= len(m.manifest.Files) {
			m.done = true
			return m, tea.Quit
		}

		url := fmt.Sprintf(baseURL, m.manifest.Files[m.index].ProjectID, m.manifest.Files[m.index].FileID)
		cmd, name, err := downloadFile(downloadPath, url, m.manifest.Files[m.index].Name)
		if err != nil && err.Error() != "mod not found" {
			log.Fatal(err)
		}
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.manifest.Files)-1))
		return m, tea.Batch(
			progressCmd,
			printer(name, err),
			cmd,
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.done {
		text := fmt.Sprintf("Done! Downloaded %d mods.\n", len(m.manifest.Files)-len(m.errors))
		if len(m.errors) > 0 {
			text += "Errors: "
			for _, n := range m.errors {
				text += n + ", "
			}
		}
		return marginStyle.Render(text)
	}

	if m.index < len(m.manifest.Files) {
		n := len(m.manifest.Files)
		w := lipgloss.Width(fmt.Sprintf("%d", n))

		count := fmt.Sprintf(" %*d/%*d", w, m.index, w, n-1)

		spin := m.spinner.View() + "  "
		prog := m.progress.View()

		info := "Downloading " + glowStyle.Render(m.manifest.Files[m.index].Name)

		cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+count))
		gap := strings.Repeat(" ", cellsRemaining)

		out := spin + info + gap + prog + count
		return out
	}
	return ""
}

func printer(name string, err error) tea.Cmd {
	if err != nil {
		return tea.Printf("%s%s", crossStyle, errorStyle.Render(name))
	}
	return tea.Printf("%s%s", checkStyle, downloadedStyle.Render(name))
}

func main() {
	body, err := os.ReadFile("./manifest.json")
	if err != nil {
		log.Fatal("Error when opening file:", err)
	}
	var manifest Manifest
	if err = json.Unmarshal(body, &manifest); err != nil {
		log.Fatal("Error during manifest unmarshal:", err)
	}

	names, err := getModNames()
	if err != nil {
		log.Println("Error during modlist reading:", err)
	}

	for i, el := range manifest.Files {
		manifest.Files[i].Name = names[strconv.Itoa(el.ProjectID)]
	}

	if _, err := tea.NewProgram(newModel(manifest)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

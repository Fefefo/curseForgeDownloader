package main

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Manifest struct {
	Minecraft Minecraft `json:"minecraft"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Author    string    `json:"author"`
	Files     []File    `json:"files"`
	Overrides string    `json:"overrides"`
}

type File struct {
	ProjectID int  `json:"projectID"`
	FileID    int  `json:"fileID"`
	Required  bool `json:"required"`
	Name      string
}

type Minecraft struct {
	Version    string `json:"version"`
	ModLoaders []struct {
		ID      string `json:"id"`
		Primary bool   `json:"primary"`
	} `json:"modLoaders"`
}

var (
	baseURL = "https://www.curseforge.com/api/v1/mods/%d/files/%d/download"
	//secondaryURL = "https://www.curseforce.com/projects/%d"
)

func downloadFile(filepath string, url string, nameNoVersion string) (tea.Cmd, string, error) {
	time.Sleep(5 * time.Millisecond)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	redirectedURL := resp.Header.Get("Location")
	splitURL := strings.Split(redirectedURL, "/")
	name := strings.Split(splitURL[len(splitURL)-1], "?")[0]

	out, errMod := os.Create(fmt.Sprintf(filepath, name))
	if errMod != nil {
		return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return modDownloaded(" ")
		}), nameNoVersion, errors.New("mod not found")
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, "", err
	}

	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return modDownloaded(" ")
	}), name, err
}

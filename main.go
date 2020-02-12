package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type SteamGame struct {
	ID int
	Name, Dir, Size string
	Library uint8
}

func (game *SteamGame) FormatSize() string {
	// First convert to int
	bytes, err := strconv.Atoi(game.Size)
	if err != nil {
		return "0 b"
	}
	// gb
	if bytes > 1000000000 {
		return fmt.Sprintf("%v gb", bytes / 1000000000)
	}
	// mb
	return fmt.Sprintf("%v mb", bytes / 1000000)
}

func (game *SteamGame) FullPath() string {
	return path.Join(libraries[game.Library], "common", game.Dir)
}

var libraries []string

var games map[int]SteamGame

// Simple VDF/ACF parser, doesn't handle objects/arrays etc.
func ParseVDF(data string) map[string]string {
	// Get the actual file content
	start := strings.Index(data, "{")
	end   := strings.LastIndex(data, "}")
	if start < 0 || end < 0 {
		fmt.Printf("warning: start/end not found (%v/%v)\n", start, end)
		return nil
	}
	lines := strings.Split(data[start+1:end], "\n")
	// Loop through all lines
	mapData := make(map[string]string)
	for _, line := range lines {
		// Remove all quotes
		line = strings.ReplaceAll(strings.TrimSpace(line), "\"", "")
		// Separate by double tab
		l := strings.Split(line, "\t\t")
		// Insert in map
		if len(l) != 2 {
			continue
		}
		mapData[l[0]] = l[1]
	}
	// Return final map
	return mapData
}

func Refresh() {
	homeDir, _ := os.UserHomeDir()
	libPath := path.Join(homeDir, ".local/share/Steam/steamapps/")
	libData, _ := ioutil.ReadFile(path.Join(libPath, "libraryfolders.vdf"))
	games = make(map[int]SteamGame)
	libraries = make([]string, 1)
	libraries[0] = libPath

	// Get paths from file
	for key, value := range ParseVDF(string(libData)) {
		// See if key is a number
		_, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		libraries = append(libraries, path.Join(value, "steamapps"))
	}

	// Get all games in the libraries
	for libIndex, lib := range libraries {
		dir, _ := ioutil.ReadDir(lib)
		// Check all files in directory
		for _, file := range dir {
			// See if it's a manifest file
			if !strings.HasPrefix(file.Name(), "appmanifest_") || !strings.HasSuffix(file.Name(), ".acf") {
				continue
			}
			// Parse it
			appContent, _ := ioutil.ReadFile(path.Join(lib, file.Name()))
			appData := ParseVDF(string(appContent))
			appID, _ := strconv.Atoi(appData["appid"])
			// Add it to games index
			games[appID] = SteamGame{
				ID:      appID,
				Name:    appData["name"],
				Library: uint8(libIndex),
				Dir:     appData["installdir"],
				Size:    appData["SizeOnDisk"],
			}
		}
	}
}

func Search(keyword string) *SteamGame {
	// First check if it's an id
	id, err := strconv.Atoi(keyword)
	if err == nil {
		game, found := games[id]
		if found {
			return &game
		}
	}
	// Otherwise, loop through and check name
	for _, game := range games {
		if strings.Contains(strings.ToLower(game.Name), keyword) {
			return &game
		}
	}
	// Nothing found
	return nil
}

func Uninstall(game *SteamGame) error {
	// Start by deleting main folder
	// We could just do os.RemoveAll, but we don't get any output :(
	err := filepath.Walk(game.FullPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		if err = os.Remove(path); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Final script to let Steam know it's installed
	manifestPath := path.Join(libraries[game.Library], fmt.Sprintf("appmanifest_%v.acf", game.ID))
	fmt.Println(manifestPath)
	return os.Remove(manifestPath)
}

func main() {
	// Check correct amount of arguments
	if len(os.Args) != 2 {
		fmt.Println("usage: sgu <id/name>")
		return
	}
	// Refresh cache
	Refresh()
	// Search for game
	game := Search(os.Args[1])
	if game == nil {
		fmt.Println("no results found for", os.Args[1])
		return
	}
	fmt.Printf("%v\n\t%v (%v)\nuninstall? [y/n]: ", game.FullPath(), game.Name, game.FormatSize())
	var response string
	_, _ = fmt.Scanf("%s", &response)
	if response == "y" || response == "Y" {
		if err := Uninstall(game); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: failed to uninstall \"%v\": %v", game.Name, err)
		}
	}
}
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    choices  []string           // items on the to-do list
    cursor   int                // which to-do list item our cursor is pointing at
    selected map[int]struct{}   // which to-do items are selected
}

func initialModel() model {
	return model{
		// Our to-do list is a grocery list
		choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
}

func main() {
	// Define the directory path to search within
	dir := filepath.Join(os.Getenv("HOME"), "Projects")

	startSessFileList, err := findAllStartSessionScripts(dir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var options []string
	// get the icons for the directory
	for _, starSessFile := range startSessFileList {
		filesList, err := listFiles(filepath.Dir(starSessFile))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		icons := getIconsForDir(filesList)
		printStr := getUniqueIconString(compressIcons(icons))

		// print the directory name and the icons
		// fmt.Printf("%s: %s\n", strings.Replace(filepath.Dir(starSessFile), dir, "", 1), printStr)
		str := fmt.Sprintf("%s: %s", strings.Replace(filepath.Dir(starSessFile), dir, "", 1), printStr)
		options = append(options, str)
	}

  // #### Bubble Tea ####
  
  // init the model
}

func getUniqueIconString(icons map[IconDetails]int) string {
	var iconString string
	includedIcons := make(map[string]bool)
	for icon := range icons {
		if _, ok := includedIcons[icon.Icon]; !ok {
			// cprint(icon.CtermColor, icon.Icon+" ")
			iconString += cprintbuilder(icon.CtermColor, icon.Icon+" ")
			includedIcons[icon.Icon] = true
		}
	}
	return iconString
}

func getIconsForDir(fileList []string) map[IconDetails]int {
	uniqueIcons := make(map[IconDetails]int)
	for _, file := range fileList {
		// Get only the file name from the full path
		fileName := filepath.Base(file)
		icon, score := findIconForFile(fileName)
		// ignore if in ignore list
		if contains(ignoreNamesList, icon.Name) {
			continue
		}
		// add the file into the set
		uniqueIcons[icon] = score
	}

	return uniqueIcons
}

func compressIcons(icons map[IconDetails]int) map[IconDetails]int {
	// if react is found remove html, css, js
	if _, isReactPresent := icons[iconsByFileExtension["jsx"]]; isReactPresent {
		compressReact(icons)
	}

	// if tsx is found and jsx
	if _, isTsxPresent := icons[iconsByFileExtension["tsx"]]; isTsxPresent {
		compressTSX(icons)
	}

	// if gd is found remove godot, tres, tscn,
	if _, isGdPresent := icons[iconsByFileExtension["gd"]]; isGdPresent {
		compressGodot(icons)
	}
	return icons
}

func getExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}

func findIconForFile(filename string) (IconDetails, int) {
	// Define the icons for different file types
	maps := []map[string]IconDetails{
		iconsByFileName,
		iconsByFileOS,
		iconsByDesktopEnvironment,
		iconsByWindowManager,
	}

	// Iterate over the maps and return the icon if found
	for i, m := range maps {
		if val, ok := m[filename]; ok {
			return val, i
		}
	}

	// Return the icon if the file extension is found
	if val, ok := iconsByFileExtension[getExtension(filename)]; ok {
		return val, 5
	}

	// Return the default icon if no icon is found
	return DefaultIcon, 0
}

// checks if a .git folder exists within the given path
func hasGitFolder(rootPath string) (bool, error) {
	found := false

	// WalkFunc to check if the directory is .git
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			found = true
			return filepath.SkipDir // Stop walking further
		}
		return nil
	}

	// Walk through the directory tree starting from rootPath
	err := filepath.Walk(rootPath, walkFn)
	if err != nil {
		return false, err
	}

	return found, nil
}

// listFiles lists all files in a directory excluding certain directories
func listFiles(rootPath string) ([]string, error) {
	var files []string

	// WalkFunc to append files to the slice
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore directories starting with "." or named "node_modules"
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules") {
			return filepath.SkipDir
		}
		// Only add regular files to the list
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}

	// Walk through the directory tree starting from rootPath
	err := filepath.Walk(rootPath, walkFn)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// Function to search for all start-session.sh files in the given directory and its subdirectories
func findAllStartSessionScripts(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories starting with "." or named "node_modules"
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules") {
			return filepath.SkipDir
		}

		// Check if the file is named start-session.sh
		if info.Name() == "start-session.sh" {
			files = append(files, path)
		}
		return nil
	})
	// Check the error returned from Walk
	if err != nil {
		return nil, fmt.Errorf("error searching directory: %v", err)
	}

	return files, nil
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func cprintbuilder(color string, message string) string {
	return fmt.Sprintf("\033[38;5;%sm%s\033[0m", color, message)
}

func cprint(color string, message string) {
	s := cprintbuilder(color, message)
	fmt.Print(s)
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices             []string // items on the to-do list
	start_session_paths []string // the file path of start script
	cursor              int      // which to-do list item our cursor is pointing at
	selected            int
}

func initialModel(choices []string, start_session_paths []string) model {
	return model{
		choices:             choices,
		start_session_paths: start_session_paths,
		selected:            -1,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			m.selected = m.cursor
      folderPath := filepath.Dir(m.start_session_paths[m.selected])
      
      // open the terminal with the start-session.sh script
      cmd := tea.Batch(
        tea.ExecProcess(exec.Command("cd", folderPath), nil),
        tea.ExecProcess(exec.Command("bash", m.start_session_paths[m.selected]), nil),
        tea.Quit,
      )

      return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	// The header
  s := "Select a project to initialize it: \n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
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
	model := initialModel(options, startSessFileList)

	program := tea.NewProgram(model)

	if _, err := program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
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

package services

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
    "src/config"
    "strings"
    "unicode"
)

const hiddenComponentDir = ".confetti/Components"

// Actually, there should be another letter 'c' as the first letter here,
// but we don't consider it because it can be in lowercase or uppercase.
const componentConfigSuffix = "omponent.blade.php"
const componentClassSuffix = "omponent.class.php"

func UpsertHiddenComponentE(root string, file string, verbose bool) {
	err := UpsertHiddenComponent(root, file, verbose)
	if err != nil {
		println("Err UpsertHiddenComponentE:")
		println(err.Error())
		return
	}
}

func UpsertHiddenComponent(root string, file string, verbose bool) error {
	originFile := file
	// Check if it is a component generator
	if !strings.HasSuffix(file, componentConfigSuffix) {
		if !strings.HasSuffix(file, componentClassSuffix) {
			return nil
		}
		// If composer class has changed, handle it the same as the config file
		file = strings.Replace(file, componentClassSuffix, componentConfigSuffix, 1)
	}
	if verbose {
		println("Hidden component triggered by: " + originFile)
	}
	// Get content of component
	body, err := Send("http://api.localhost/parser/source/component?file=/"+file, nil, http.MethodGet)
	if err != nil {
		return err
	}
	// Get file content from response
	contentRaw := map[string]string{}
	json.Unmarshal([]byte(body), &contentRaw)
	content64 := contentRaw["content"]
	name := contentRaw["name_class"]
	content, err := base64.StdEncoding.DecodeString(content64)
	if err != nil {
		return err
	}
	// Save hidden component
	target := path.Join(root, hiddenComponentDir, name+".php")
	err = os.MkdirAll(path.Dir(target), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(string(content))
	if verbose {
		println("Hidden component saved: " + target)
	}
	return err
}

func UpsertHiddenMap(root string, verbose bool) error {
	// Compose hidden Map component
	names, err := getComponentClassNamesByDirectory(path.Join(root, hiddenComponentDir))
	if err != nil {
		return err
	}
    content := getMapContent(names)
	// Save hidden Map component
	target := path.Join(root, hiddenComponentDir, "Map.php")
	err = os.MkdirAll(path.Dir(target), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
    _, err = f.WriteString(content)
	if verbose {
		println("Hidden component saved: " + target)
	}
	return err
}

func getComponentClassNamesByDirectory(dir string) ([]string, error) {
	result := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// We want to generate the map class, so ignore it
		if file.Name() == "Map.php" {
			continue
		}
        // We assume that the filename is equal to the classname.
		name := strings.TrimRight(file.Name(), ".php")
		result = append(result, name)
	}
	return result, nil
}

func getMapContent(classNames []string) string {
	contentRaw, err := config.Embed.Template.ReadFile("Map.php")
    if err != nil {
        panic(err)
    }
    content := string(contentRaw)
	for _, className := range classNames {
        functionName := lowerFirst(className)
		function := `public function ` + functionName + `(string $key): ` + className + `
    {
        return new ` + className + `();
    }
//-> function`
        content = strings.Replace(content, "//-> function", function, 1)
	}
    content = strings.Replace(content, "//-> function", "", 1)
    return content
}

func lowerFirst(input string) string {
    if len(input) == 0 {
        return ""
    }
    char := []rune(input)
    char[0] = unicode.ToLower(char[0])
    return string(char)
}

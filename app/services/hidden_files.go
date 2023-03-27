package services

import (
    "encoding/base64"
    "encoding/json"
    "net/http"
    "os"
    "path"
    "strings"
)

const hiddenComponentDir = ".confetti/components"
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
    body, err := Send("http://api.localhost/parser/source/component?file=/" + file, nil, http.MethodGet)
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
    target := path.Join(root, hiddenComponentDir, name + ".php")
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
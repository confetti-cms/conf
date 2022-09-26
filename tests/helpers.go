package tests

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"src/app/services"
	"strings"

	"github.com/spf13/cast"
)

const mockDir = "mock_generated"

func initTestGit() string {
	pc, _, _, _ := runtime.Caller(1)
	testDir := strings.Split(runtime.FuncForPC(pc).Name(), ".")[1]
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	dir := path.Join(currentDir, mockDir, testDir)
	// Clean up directory from old test
	_, err = os.Stat(dir)
	if !os.IsNotExist(err) {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Create directory
	err = os.Mkdir(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	// Create empty git repository
	cmd := exec.Command("git", "init", dir)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to run `git init`: %s", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func touchFile(dir, fileName string) {
	// Ensure dir exists
	subDir := path.Dir(fileName)
	fullDir := filepath.Join(dir, subDir)
	os.MkdirAll(fullDir, 0755)
	// Create file
	file, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}

func gitAdd(dir, file string) {
	// Add file or directory to the stage
	_, err := services.GitAdd(filepath.Join(dir, file))
	if err != nil {
		log.Fatalf("failed to run `git add`: %s", err)
	}
}

func gitCommit(dir, file string) {
	// Commit file
	_, err := services.GitCommit(filepath.Join(dir, file))
	if err != nil {
		log.Fatalf("failed to run `git commit`: %s", err)
	}
}

func setFileContent(dir, fileName, content string) {
	file, err := os.OpenFile(filepath.Join(dir, fileName), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
}

func deleteFile(dir, fileName string) {
	err := os.Remove(filepath.Join(dir, fileName))
	if err != nil {
		log.Fatal(err)
	}
}

func getCommitFromLog(dir string, since int) string {
	raw, err := services.RunCommand(fmt.Sprintf(`cd %s && git rev-parse HEAD~%d`, dir, since))
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatal(cast.ToString(ee.Stderr))
		}
		log.Fatal(err)
	}
	if strings.Contains(raw, "fatal") {
		log.Fatal(raw)
	}
	return strings.Trim(raw, "\n")
}

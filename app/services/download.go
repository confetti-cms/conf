package services

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"src/config"
	"strings"
	"time"

	"github.com/confetti-framework/framework/inter"
)

func DownloadZip(cli inter.Cli, url string, outputPath string, env Environment) error {
	token, err := GetAccessToken(cli, env)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if config.App.VeryVeryVerbose {
		fmt.Printf("Method: %s, URL: %s\n", "GET", url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Perform the request
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return fmt.Errorf("error with status: %d while downloading from: %s", res.StatusCode, url)
	}

	// Create a temporary file to store the downloaded zip
	tempFile, err := os.CreateTemp("", "download-*.zip")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure the temp file is removed after use
	defer tempFile.Close()

	// Write the response body to the temp file
	_, err = io.Copy(tempFile, res.Body)
	if err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}

	// Unzip the file to the outputPath
	err = Unzip(tempFile.Name(), outputPath)
	if err != nil {
		return fmt.Errorf("error unzipping file %s: %w", tempFile.Name(), err)
	}

	fmt.Printf("File downloaded and unzipped successfully to %s\n", outputPath)
	return nil
}

// Unzip extracts a zip file to the specified destination directory
func Unzip(src string, dest string) error {
	// print the first 10 lines of the zip file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", src, err)
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat zip file %s: %w", src, err)
	}
	if fi.IsDir() {
		return fmt.Errorf("zip file %s is a directory", src)
	}
	// Check if the destination directory exists, if not create it
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.MkdirAll(dest, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create destination directory %s: %w", dest, err)
		}
	}
	println("Unzipping file", src, "to", dest)

	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", src, err)
	}
	defer r.Close()

	for _, f := range r.File {
		fPath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fPath)
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fPath, err)
			}
			continue
		}

		err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create parent directories for %s: %w", fPath, err)
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip file %s: %w", f.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to copy contents to file %s: %w", fPath, err)
		}
	}
	return nil
}

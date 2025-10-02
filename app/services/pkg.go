package services

import (
	"fmt"
	"os"
	"path/filepath"
	"src/config"
	"strings"

	"github.com/confetti-framework/framework/support"
)

func HasModifications() ([]string, error) {
	cmd := fmt.Sprintf(`cd %s && git diff --name-only && git diff --name-only --cached | sort -u`, config.Path.Root)
	out, err := RunCommand(cmd)
	if err != nil {
		return nil, err
	}
	if out == "" {
		if config.App.Verbose {
			fmt.Println("No modifications found.")
		}
		return nil, nil
	}

	// Split the output into lines and count them
	lines := strings.FieldsFunc(string(out), func(r rune) bool { return r == '\n' })
	return lines, nil
}

func RestoreDirectory(pkg string) (bool, error) {
	// First we check if the directory has been created in the past.
	cmd := fmt.Sprintf("cd %s && git --no-pager log -1 --format=\"%%H\" -- pkg/%s", config.Path.Root, pkg)
	hash, err := RunCommand(cmd)
	if err != nil {
		support.Dump("Error checking if package directory exists %w, command output: %s", err, hash)
		return false, fmt.Errorf("error checking if package %s exists: %w", pkg, err)
	}

	hash = strings.TrimSpace(hash)
	if hash == "" {
		if config.App.Verbose {
			fmt.Printf("Package %s has not been created in the past, adding new package...\n", pkg)
		}
		return false, nil
	}

	if config.App.Verbose {
		fmt.Printf("Package %s has been created in the past, restoring it with hash %s...\n", pkg, hash)
	}

	// If the directory exists, we restore it to the last commit.
	cmd = fmt.Sprintf("cd %s && git checkout %s^ -- pkg/%s", config.Path.Root, hash, pkg)
	_, err = RunCommand(cmd)
	if err != nil {
		return false, fmt.Errorf("error checking out package %s with hash %s: %w", pkg, hash, err)
	}

	// Commit the restoration of the package directory.
	err = CommitChanges(pkg, fmt.Sprintf("Restore package %s", pkg))
	if err != nil {
		return false, fmt.Errorf("error committing restoration of package %s: %w", pkg, err)
	}
	return true, nil
}

func AddNewPackage(pkg string) error {
	cmd := fmt.Sprintf("cd %s && git subtree add --prefix=\"pkg/%s\" git@github.com:%s.git main", config.Path.Root, pkg, pkg)
	output, err := RunCommand(cmd)
	if err != nil {
		// Check if the error is due to branch not existing
		if strings.Contains(output, "doesn't exist") || strings.Contains(output, "not found") {
			fmt.Printf("\n\033[31mThe branch 'main' does not exist in repository '%s'.\n\033[0m", pkg)
			fmt.Println("Please create the repository and initial commit first. You can use the following commands to initialize it:")
			fmt.Println()
			fmt.Printf("\033[32mcd /path/to/%s\n\033[0m", pkg)
			fmt.Printf("\033[32mecho \"# %s\" >> README.md\n\033[0m", pkg)
			fmt.Printf("\033[32mgit init\n\033[0m")
			fmt.Printf("\033[32mgit add README.md\n\033[0m")
			fmt.Printf("\033[32mgit commit -m \"first commit\"\n\033[0m")
			fmt.Printf("\033[32mgit branch -M main\n\033[0m")
			fmt.Printf("\033[32mgit remote add origin git@github.com:%s.git\n\033[0m", pkg)
			fmt.Printf("\033[32mgit push -u origin main\n\033[0m")
			fmt.Printf("\033[32mcd -\n\033[0m")
			fmt.Println()
			return fmt.Errorf("branch 'main' does not exist in repository %s, please create it first", pkg)
		}
		return fmt.Errorf("error adding new package %s: %w", pkg, err)
	}
	return nil
}

func PullLatestChanges(pkg string) error {
	cmd := fmt.Sprintf("cd %s && git subtree pull --message=\"Pull package %s\" --prefix=\"pkg/%s\" git@github.com:%s.git main", config.Path.Root, pkg, pkg, pkg)
	// We can't use StreamCommand here (for now) because the command gives exit code 1 (if there are no changes).
	output, err := RunCommand(cmd)
	if err != nil {
		// Check if the error is due to repository not existing
		if strings.Contains(output, "Repository not found") || strings.Contains(output, "does not exist") {
			fmt.Printf("\n\033[31mThe repository '%s' does not exist.\n\033[0m", pkg)
			fmt.Println("Please create the repository first. You can use the following commands to initialize it:")
			fmt.Println()
			fmt.Printf("\033[32mcd /path/to/%s\n\033[0m", pkg)
			fmt.Printf("\033[32mecho \"# %s\" >> README.md\n\033[0m", pkg)
			fmt.Printf("\033[32mgit init\n\033[0m")
			fmt.Printf("\033[32mgit add README.md\n\033[0m")
			fmt.Printf("\033[32mgit commit -m \"first commit\"\n\033[0m")
			fmt.Printf("\033[32mgit branch -M main\n\033[0m")
			fmt.Printf("\033[32mgit remote add origin git@github.com:%s.git\n\033[0m", pkg)
			fmt.Printf("\033[32mgit push -u origin main\n\033[0m")
			fmt.Printf("\033[32mcd -\n\033[0m")
			fmt.Println()
			return fmt.Errorf("repository %s does not exist, please create it first", pkg)
		}
		return err
	}
	return nil
}

func PushPackage(pkg string) error {
	cmd := fmt.Sprintf("cd %s && git subtree push --prefix=\"pkg/%s\" git@github.com:%s.git main", config.Path.Root, pkg, pkg)
	_, err := StreamCommand(cmd)
	if err != nil {
		return err
	}
	return nil
}

func PkgPackageContainsPhpFiles(pkg string) bool {
	pkgPath := filepath.Join(config.Path.Root, "pkg", pkg)
	if config.App.Verbose {
		fmt.Printf("Checking if package %s contains PHP files in directory: %s\n", pkg, pkgPath)
	}

	found := false
	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".php") {
			found = true
			if config.App.VeryVerbose {
				fmt.Printf("Found PHP file: %s\n", path)
			}
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		support.Dump("Error checking for PHP files in package %s: %v", pkg, err)
	}

	return found
}

func UpdateComposer(pkg string) error {
	cmd := fmt.Sprintf("cd %s && composer require %s \"*\" --ignore-platform-reqs", config.Path.Root, pkg)
	_, err := RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error updating composer for package %s: %w", pkg, err)
	}
	return nil
}

func RemovePackage(pkg, msg string) error {
	err := os.RemoveAll(filepath.Join("pkg", pkg))
	if err != nil {
		return fmt.Errorf("error removing package %s: %w", pkg, err)
	}
	// only commit the pkg/package directory
	cmd := fmt.Sprintf("cd %s && git commit -m \"%s\" -- pkg/%s || true", config.Path.Root, msg, pkg)
	if config.App.Verbose {
		fmt.Printf("Running command to commit removal of package %s: %s\n", pkg, cmd)
	}
	_, err = RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error committing removal of package %s: %w", pkg, err)
	}
	return err
}

func CommitChanges(pkg, msg string) error {
	cmd := fmt.Sprintf("cd %s && git commit -am \"%s\" || true", config.Path.Root, msg)
	_, err := RunCommand(cmd)
	return err
}

func PrintPackageNoComposerMessage(pkg string) {
	fmt.Println("\n\033[31mThe package", pkg, "does not contain a composer.json file.\n\033[0m")
	fmt.Println("Please check the package directory and ensure it contains a composer.json file.")
	fmt.Printf("Example composer.json for package %s:\n", pkg)
	fmt.Printf(
		"\n\033[34m{\n"+
			"    \033[36m\"name\"\033[34m: \"\033[32m%s\033[34m\",\n"+
			"    \033[36m\"autoload\"\033[34m: \033[34m{\n"+
			"        \033[36m\"psr-4\"\033[34m: \033[34m{\n"+
			"            \"\033[32m%s\033[34m\": \"\"\n"+
			"        }\n"+
			"    }\n"+
			"}\033[0m\n\n",
		pkg, namespaceExampleFromPackage(pkg),
	)

}

func PrintPackageInstalledMessage(pkg string) {
	fmt.Println("\n\033[32mThe package", pkg, "is now installed 🎁\n\033[0m")
	fmt.Println("The package", pkg, "is fully integrated into your repository as if you wrote it yourself. You can make changes directly in your repository, and the commits will be saved to your own history. It’s perfectly fine to maintain a customized version of the package within your main project.")
	fmt.Println("If you have sufficient permissions, you can push your changes to the central package repository using: \033[34mconfetti pkg:push " + pkg + "\033[0m")
}

func PrintPackagePulledMessage(pkg string) {
	fmt.Println("\n\033[32mThe package", pkg, "has been pulled successfully 🧁\n\033[0m")
}

func namespaceExampleFromPackage(pkg string) string {
	splitted := strings.Split(pkg, "/")
	repo := toPascalCase(splitted[0])
	name := toPascalCase(splitted[1])

	return fmt.Sprintf("%s\\\\%s\\\\", repo, name)
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
}

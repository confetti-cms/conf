package services

import (
	"fmt"
	"os"
	"path/filepath"
	"src/config"
	"strings"

	"github.com/confetti-framework/framework/support"
)

//     "pkg:pull": [
//   "sh -c 'echo \"\\n...Checking for modifications...\"' -",
//   "sh -c 'git diff-index --quiet HEAD -- || (echo \"\\nYour project has modifications, please commit and push or stash them first.\n\" && exit 1)' -",
//   "sh -c '[ -d \"pkg/$1\" ] && true || ((git --no-pager log -1 --format=\"%H\" -- pkg/$1 | grep -q .) && (echo -e \"\\n...Trying to restore directory...\\n\" && git checkout $(git --no-pager log -1 --format=\"%H\" -- pkg/$1)^ -- pkg/$1 && git commit -am \"Restore package $1\") || true;)' -",
//   "sh -c '[ -d \"pkg/$1\" ] && true || ((echo \"\\n...Not created in the past, add new package...\\n\" && git subtree add --prefix=\"pkg/$1\" git@github.com:$1.git main);)' -",
//   "sh -c 'echo \"\\n...Pull latest changes...\\n\"' -",
//   "sh -c 'git subtree pull --message=\"Pull package $1\" --prefix=\"pkg/$1\" git@github.com:$1.git main' -",
//   "sh -c 'echo \"\\n...Updating composer.json and add the package to the autoloader...\\n\"' -",
//   "sh -c 'composer require $1 \"*\" --ignore-platform-reqs' -",
//   "sh -c 'git commit -am \"Added package to composer $1\" || true' -",
//   "sh -c 'echo \"\\n...The package $1 is now installed 🎁\\n\"' -",
//   "sh -c 'echo \"The package $1 is fully integrated into your repository as if you wrote it yourself.\"' -",
//   "sh -c 'echo \"You can make changes directly in your repository, and the commits will be saved to your own history.\"' -",
//   "sh -c 'echo \"It’s perfectly fine to maintain a customized version of the package within your main project.\"' -",
//   "sh -c 'echo \"If you have sufficient permissions, you can push your changes to the central package repository using: `confetti pkg:push $1`.\"' -"
//
// Make for this seperate functions
// 	cmd := fmt.Sprintf("cd %s && composer install --ignore-platform-reqs --no-interaction --no-progress --no-plugins", config.Path.Root)
// Ignore the result because composer's warnings go to stderr.
// err = StreamCommand(cmd)
//
// This has to be good for windows and linux

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
	_, err := RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error adding new package %s: %w", pkg, err)
	}
	return nil
}

func PullLatestChanges(pkg string) error {
	cmd := fmt.Sprintf("cd %s && git subtree pull --message=\"Pull package %s\" --prefix=\"pkg/%s\" git@github.com:%s.git main", config.Path.Root, pkg, pkg, pkg)
	_, err := RunCommand(cmd)
	if err != nil {
		return err
	}
	return nil
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
		"\n\033[34m{\n\n"+
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

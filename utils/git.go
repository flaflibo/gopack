package utils

import (
	"fmt"
	"os"
	"os/exec"
)

type Git struct {
}

/**
* Checkout a specific branch / tag
* branch: the name of the branch / tag
* repoPath: the path to the git repositroy on which to execute checkout command
 */
func (g *Git) Checkout(branch, repoPath string) error {
	checkoutCmd := []string{"checkout", branch}
	cmd := exec.Command("git", checkoutCmd...)
	cmd.Dir = repoPath
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

/**
* Returns the current checkedout branch
 */
func (g *Git) GetCurrentBranch(repoPath string) (string, error) {
	currentBranchCmd := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	cmd := exec.Command("git", currentBranchCmd...)
	cmd.Dir = repoPath
	currentBranch, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(currentBranch), nil
}

/*
* Clone a git repo
* url: the url from where to clone the repo
* out: the output path where to store the repo as subdirectory
 */
func (g *Git) Clone(url string, out string) error {
	cloneCmd := []string{"clone", url}
	cmd := exec.Command("git", cloneCmd...)
	cmd.Dir = out
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

/*
* Execute Pull for the repository specify by repoPath
 */
func (g *Git) Pull(repoPath string) error {
	pullCmd := []string{"pull"}
	cmd := exec.Command("git", pullCmd...)
	cmd.Dir = repoPath
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

/*
* clone a git repositroy. Repositry will be cloned into the out
* path
* name: name of the repo
* branch: the branch to checkout
* out: output path
 */
func (g *Git) CloneAndCheckout(name, branch, url, out string) error {
	//check if repo already exists
	repoExists := false
	repoPath := fmt.Sprintf("%s/%s", out, name)

	repoInfo, err := os.Stat(repoPath)
	if err != nil {
		repoExists = false
	} else {
		repoExists = repoInfo.IsDir()
	}

	if repoExists {
		//The repo already exists so check if we have the correct branch checked out
		currentBranch, err := g.GetCurrentBranch(repoPath)
		if err != nil {
			return err
		}

		if string(currentBranch) != branch {
			//we do not have the correct branch checked out so do this now
			if err := g.Checkout(branch, out); err != nil {
				return err
			}
		}

		//now pull the latest changes for that branch
		if err := g.Pull(repoPath); err != nil {
			return err
		}

	} else {
		//repo does not exists
		if err := g.Clone(url, out); err != nil {
			return err
		}

		if err := g.Checkout(branch, repoPath); err != nil {
			return err
		}
	}
	return nil
}

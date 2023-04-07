package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type DecodedSymmetricEncryptionDetails = struct {
	Cipher []byte
	IV     []byte
	Tag    []byte
	Key    []byte
}

func GetBase64DecodedSymmetricEncryptionDetails(key string, cipher string, IV string, tag string) (DecodedSymmetricEncryptionDetails, error) {
	cipherx, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return DecodedSymmetricEncryptionDetails{}, fmt.Errorf("Base64DecodeSymmetricEncryptionDetails: Unable to decode cipher text [err=%v]", err)
	}

	keyx, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return DecodedSymmetricEncryptionDetails{}, fmt.Errorf("Base64DecodeSymmetricEncryptionDetails: Unable to decode key [err=%v]", err)
	}

	IVx, err := base64.StdEncoding.DecodeString(IV)
	if err != nil {
		return DecodedSymmetricEncryptionDetails{}, fmt.Errorf("Base64DecodeSymmetricEncryptionDetails: Unable to decode IV [err=%v]", err)
	}

	tagx, err := base64.StdEncoding.DecodeString(tag)
	if err != nil {
		return DecodedSymmetricEncryptionDetails{}, fmt.Errorf("Base64DecodeSymmetricEncryptionDetails: Unable to decode tag [err=%v]", err)
	}

	return DecodedSymmetricEncryptionDetails{
		Key:    keyx,
		Cipher: cipherx,
		IV:     IVx,
		Tag:    tagx,
	}, nil
}

func IsSecretEnvironmentValid(env string) bool {
	if env == "prod" || env == "dev" || env == "test" || env == "staging" {
		return true
	}
	return false
}

func IsSecretTypeValid(s string) bool {
	if s == "personal" || s == "shared" {
		return true
	}
	return false
}

func RequireLogin() {
	currentUserDetails, err := GetCurrentLoggedInUserDetails()

	if err != nil {
		HandleError(err, "unable to retrieve your login details")
	}

	if !currentUserDetails.IsUserLoggedIn {
		PrintErrorMessageAndExit("You must be logged in to run this command. To login, run [infisical login]")
	}

	if currentUserDetails.LoginExpired {
		PrintErrorMessageAndExit("Your login expired, please login in again. To login, run [infisical login]")
	}

	if currentUserDetails.UserCredentials.Email == "" && currentUserDetails.UserCredentials.JTWToken == "" && currentUserDetails.UserCredentials.PrivateKey == "" {
		PrintErrorMessageAndExit("One or more of your login details is empty. Please try logging in again via by running [infisical login]")
	}
}

func RequireServiceToken() {
	serviceToken := os.Getenv(INFISICAL_TOKEN_NAME)
	if serviceToken == "" {
		PrintErrorMessageAndExit("No service token is found in your terminal")
	}
}

func RequireLocalWorkspaceFile() {
	workspaceFilePath, _ := FindWorkspaceConfigFile()
	if workspaceFilePath == "" {
		PrintErrorMessageAndExit("It looks you have not yet connected this project to Infisical", "To do so, run [infisical init] then run your command again")
	}

	workspaceFile, err := GetWorkSpaceFromFile()
	if err != nil {
		HandleError(err, "Unable to read your project configuration, please try initializing this project again.", "Run [infisical init]")
	}

	if workspaceFile.WorkspaceId == "" {
		PrintErrorMessageAndExit("Your project id is missing in your local config file. Please add it or run again [infisical init]")
	}
}

func GetHashFromStringList(list []string) string {
	hash := sha256.New()

	for _, item := range list {
		hash.Write([]byte(item))
	}

	sum := sha256.Sum256(hash.Sum(nil))
	return fmt.Sprintf("%x", sum)
}

// execCmd is a struct that holds the command and arguments to be executed.
// By using this struct, we can easily mock the command and arguments.
type execCmd struct {
	cmd  string
	args []string
}

var getCurrentBranchCmd = execCmd{
	cmd:  "git",
	args: []string{"symbolic-ref", "--short", "HEAD"},
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command(getCurrentBranchCmd.cmd, getCurrentBranchCmd.args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return path.Base(strings.TrimSpace(out.String())), nil
}

func GetSplitPathByDash(path string) []string {
	pathParts := strings.Split(path, "/")
	var filteredPathParts []string
	for _, s := range pathParts {
		if s != "" {
			filteredPathParts = append(filteredPathParts, s)
		}
	}

	return filteredPathParts
}

// NormalizePath cleans up a path by removing empty parts, duplicate slashes,
// and ensuring it starts with ROOT_FOLDER_PATH.
func NormalizePath(path string) string {
	ROOT_FOLDER_PATH := "/"

	if path == "" || path == ROOT_FOLDER_PATH {
		return ROOT_FOLDER_PATH
	}

	pathParts := strings.Split(path, "/")
	nonEmptyParts := []string{}
	for _, part := range pathParts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}

	cleanPathString := ROOT_FOLDER_PATH + strings.Join(nonEmptyParts, "/")
	return cleanPathString
}

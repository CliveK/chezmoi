package chezmoi

// FIXME this file needs a proper review in the morning

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/multierr"
)

// GPGEncyptionTool interfaces with gpg.
type GPGEncyptionTool struct {
	Command   string
	Recipient string
	Symmetric bool
}

// Decrypt implements EncyptionTool.Decrypt.
func (g *GPGEncyptionTool) Decrypt(filenameHint string, ciphertext []byte) (plaintext []byte, err error) {
	filename, cleanup, err := g.DecryptToFile(filenameHint, ciphertext)
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, cleanup())
	}()
	return ioutil.ReadFile(filename)
}

// DecryptToFile implements EncyptionTool.DecryptToFile.
func (g *GPGEncyptionTool) DecryptToFile(filenameHint string, ciphertext []byte) (filename string, cleanupFunc CleanupFunc, err error) {
	cleanupFunc = nullCleanupFunc

	tempDir, err := ioutil.TempDir("", "chezmoi-gpg-decrypt")
	if err != nil {
		return
	}
	cleanupFunc = func() error {
		return os.RemoveAll(tempDir)
	}

	filename = filepath.Join(tempDir, filepath.Base(filenameHint))
	inputFilename := filename + ".gpg"
	if err = ioutil.WriteFile(inputFilename, ciphertext, 0o600); err != nil {
		return
	}

	//nolint:gosec
	cmd := exec.Command(
		g.Command,
		"--output", filename,
		"--quiet",
		"--decrypt", inputFilename,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return
}

// Encrypt implements EncryptionTool.Encypt.
func (g *GPGEncyptionTool) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	tempFile, err := ioutil.TempFile("", "chezmoi-gpg-encrypt")
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, os.RemoveAll(tempFile.Name()))
	}()

	if err = tempFile.Chmod(0o600); err != nil {
		return
	}

	if err = ioutil.WriteFile(tempFile.Name(), ciphertext, 0o600); err != nil {
		return
	}

	return g.EncryptFile(tempFile.Name())
}

// EncryptFile implements EncryptionTool.EncryptFile.
func (g *GPGEncyptionTool) EncryptFile(filename string) (ciphertext []byte, err error) {
	outputFilename := filename + ".gpg"

	args := []string{
		"--armor",
		"--output", outputFilename,
		"--quiet",
	}
	if g.Symmetric {
		args = append(args, "--symmetric")
	} else {
		if g.Recipient != "" {
			args = append(args, "--recipient", g.Recipient)
		}
		args = append(args, "--encrypt")
	}
	args = append(args, filename)

	//nolint:gosec
	cmd := exec.Command(g.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return
	}

	ciphertext, err = ioutil.ReadFile(outputFilename)
	return
}

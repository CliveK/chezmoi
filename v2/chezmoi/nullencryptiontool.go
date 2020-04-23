package chezmoi

import "io/ioutil"

// A nullEncryptionTool returns its input unchanged.
type nullEncryptionTool struct{}

func (*nullEncryptionTool) Decrypt(ciphertext []byte) ([]byte, error) { return ciphertext, nil }
func (*nullEncryptionTool) DecryptToFile(filenameHint string, ciphertext []byte) (string, CleanupFunc, error) {
	return filenameHint, nullCleanupFunc, nil
}
func (*nullEncryptionTool) Encrypt(plaintext []byte) ([]byte, error) { return plaintext, nil }
func (*nullEncryptionTool) EncryptFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// nullCleanupFunc does nothing.
func nullCleanupFunc() error { return nil }

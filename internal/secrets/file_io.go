package secrets

import "os"

func openReadOnly(path string) (*os.File, error) {
	return os.Open(path)
}

func WrapFileInPlace(path, passphrase string) error {
	plain, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	wrapped, err := WrapWithPassphrase(plain, passphrase)
	if err != nil {
		return err
	}
	return os.WriteFile(path, wrapped, 0o600)
}

func UnwrapFileToPath(srcPath, dstPath, passphrase string) error {
	wrapped, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	plain, err := UnwrapWithPassphrase(wrapped, passphrase)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, plain, 0o600)
}

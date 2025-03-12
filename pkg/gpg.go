package pkg

import (
	"fmt"
	"os/exec"
)

// Decrypt obj
func DecryptFile(inputFile, outputFile, passphrase string) error {
	// Run the gpg command to decrypt the file
	cmd := exec.Command(
		"gpg",
		"--decrypt", // Decrypt the file
		"--batch",
		"--yes",
		"--output", outputFile, // Specify the output file
		"--passphrase", passphrase, // Passphrase for the private key
		"--pinentry-mode", "loopback", // Allow passphrase input
		inputFile, // Input file to decrypt
	)

	// Execute the command
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error decrypting file: %w\nOutput: %s", err, string(out))
	}

	//fmt.Println("File decrypted successfully:", outputFile)
	return nil
}

// Encrypt obj
func EncryptFile(inputFile, outputFile, passphrase string) error {
	// Run the gpg command to Encrypt the file
	cmd := exec.Command(
		"gpg",
		"-c", // Encrypt the file
		"--batch",
		"--yes",
		"--output", outputFile, // Specify the output file
		"--passphrase", passphrase, // Passphrase for the private key
		"--pinentry-mode", "loopback", // Allow passphrase input
		inputFile, // Input file to Encrypt
	)

	// Execute the command
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error encrypting file: %w\nOutput: %s", err, string(out))
	}

	//fmt.Println("File encrypted successfully:", outputFile)
	return nil
}

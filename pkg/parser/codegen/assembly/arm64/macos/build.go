package arm64_macos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Build assembles and links the ARM64 assembly code into an executable for macOS
func (a *arm64Macos) Build() error {
	tempDir, err := os.MkdirTemp("", "dolme_build_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	asmContent := a.GetCode()

	asmFile := filepath.Join(tempDir, "program.s")
	err = os.WriteFile(asmFile, []byte(asmContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write assembly file: %v", err)
	}

	// Assemble to object file using as
	objFile := filepath.Join(tempDir, "program.o")
	assembleCmd := exec.Command("as", "-arch", "arm64", "-o", objFile, asmFile)
	if output, err := assembleCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("assembly failed: %v\nOutput: %s", err, output)
	}

	// Link to executable using clang (more reliable than ld)
	execFile := filepath.Join(tempDir, "program")
	linkCmd := exec.Command("clang", "-arch", "arm64", "-o", execFile, objFile)
	if output, err := linkCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("linking failed: %v\nOutput: %s", err, output)
	}

	// Copy the executable to the desired output path
	copyCmd := exec.Command("cp", execFile, a.output)
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy executable: %v", err)
	}

	return nil
}

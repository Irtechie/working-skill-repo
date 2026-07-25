package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fixtureSpec struct {
	Root                string
	MutablePaths        []string
	ProofFiles          []string
	ExpectedProofHashes map[string]string
	SolutionRoot        string
	Verify              []string
	RequiredTests       []string
}

type fixtureQualification struct {
	ProofHashes   map[string]string `json:"proof_hashes"`
	BaselineRED   bool              `json:"baseline_red"`
	SolutionGREEN bool              `json:"solution_green"`
	NegativeRED   bool              `json:"negative_red"`
}

func qualifyFixture(spec fixtureSpec) (fixtureQualification, error) {
	root, err := filepath.Abs(spec.Root)
	if err != nil {
		return fixtureQualification{}, err
	}
	mutable := make(map[string]bool, len(spec.MutablePaths))
	for _, path := range spec.MutablePaths {
		clean, err := fixturePath(root, path)
		if err != nil {
			return fixtureQualification{}, err
		}
		mutable[strings.ToLower(clean)] = true
	}
	result := fixtureQualification{ProofHashes: map[string]string{}}
	for _, path := range spec.ProofFiles {
		clean, err := fixturePath(root, path)
		if err != nil {
			return fixtureQualification{}, err
		}
		if mutable[strings.ToLower(clean)] {
			return fixtureQualification{}, fmt.Errorf("proof file %q is mutable", path)
		}
		info, err := os.Lstat(clean)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fixtureQualification{}, fmt.Errorf("proof file %q must be a regular file", path)
		}
		content, err := os.ReadFile(clean)
		if err != nil {
			return fixtureQualification{}, err
		}
		sum := sha256.Sum256(content)
		slash := filepath.ToSlash(path)
		actualHash := hex.EncodeToString(sum[:])
		expectedHash := spec.ExpectedProofHashes[slash]
		if expectedHash == "" || actualHash != expectedHash {
			return fixtureQualification{}, fmt.Errorf("proof file %q hash mismatch", path)
		}
		result.ProofHashes[slash] = actualHash
	}
	if len(result.ProofHashes) == 0 {
		return fixtureQualification{}, fmt.Errorf("proof closure is empty")
	}
	if len(result.ProofHashes) != len(spec.ExpectedProofHashes) {
		return fixtureQualification{}, fmt.Errorf("proof closure hash set does not match declared files")
	}
	if err := rejectUndeclaredProofFiles(root, spec.ExpectedProofHashes); err != nil {
		return fixtureQualification{}, err
	}
	if strings.TrimSpace(spec.SolutionRoot) == "" {
		return fixtureQualification{}, fmt.Errorf("fixture executable known solution is required")
	}
	baselineRED, solutionGREEN, negativeRED, err := executeFixtureQualification(root, spec)
	if err != nil {
		return fixtureQualification{}, err
	}
	result.BaselineRED = baselineRED
	result.SolutionGREEN = solutionGREEN
	result.NegativeRED = negativeRED
	if !result.BaselineRED {
		return fixtureQualification{}, fmt.Errorf("fixture baseline is not proven RED")
	}
	if !result.SolutionGREEN {
		return fixtureQualification{}, fmt.Errorf("fixture solution is not proven GREEN")
	}
	if !result.NegativeRED {
		return fixtureQualification{}, fmt.Errorf("fixture negative mutation is not proven RED")
	}
	return result, nil
}

func verifyProofClosure(root string, expected map[string]string) error {
	for relative, expectedHash := range expected {
		path, err := fixturePath(root, relative)
		if err != nil {
			return err
		}
		actualHash, err := fileSHA256(path)
		if err != nil {
			return err
		}
		if actualHash != expectedHash {
			return fmt.Errorf("proof file %q hash mismatch", relative)
		}
	}
	return rejectUndeclaredProofFiles(root, expected)
}

func rejectUndeclaredProofFiles(root string, expected map[string]string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("fixture symlink is forbidden: %s", path)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		slash := filepath.ToSlash(relative)
		isProof := strings.HasSuffix(strings.ToLower(slash), "_test.go") ||
			strings.EqualFold(filepath.Base(slash), "go.mod") ||
			strings.EqualFold(filepath.Base(slash), "verify.test.js")
		if isProof && expected[slash] == "" {
			return fmt.Errorf("undeclared proof file %q", slash)
		}
		return nil
	})
}

func validateTaskAdmission(cfg config, task taskSpec) error {
	baseline, err := loadContextContract(cfg.ContextContracts.Baseline)
	if err != nil {
		return err
	}
	minimal, err := loadContextContract(cfg.ContextContracts.Minimal)
	if err != nil {
		return err
	}
	if err := validateContextPair(baseline, minimal); err != nil {
		return err
	}
	_, err = qualifyFixture(fixtureSpec{
		Root: task.Fixture, MutablePaths: task.MutablePaths, ProofFiles: task.ProofFiles,
		ExpectedProofHashes: task.ProofHashes, SolutionRoot: task.SolutionFixture,
		Verify: task.Verify, RequiredTests: task.RequiredTests,
	})
	return err
}

func executeFixtureQualification(root string, spec fixtureSpec) (bool, bool, bool, error) {
	tempRoot, err := os.MkdirTemp("", "amrbench-qualification-")
	if err != nil {
		return false, false, false, err
	}
	defer os.RemoveAll(tempRoot)

	baseline := filepath.Join(tempRoot, "baseline")
	if err := copyDir(root, baseline); err != nil {
		return false, false, false, err
	}
	baselineProof := runProof(baseline, spec.Verify, spec.RequiredTests)
	if baselineProof.Passed {
		return false, false, false, nil
	}

	solution := filepath.Join(tempRoot, "solution")
	if err := copyDir(root, solution); err != nil {
		return true, false, false, err
	}
	if err := applyQualificationSolution(solution, spec); err != nil {
		return true, false, false, err
	}
	solutionProof := runProof(solution, spec.Verify, spec.RequiredTests)
	if !solutionProof.Passed {
		return true, false, false, fmt.Errorf("qualification solution proof failed: %s", bounded(solutionProof.Output, 2000))
	}

	for index, mutablePath := range spec.MutablePaths {
		negative := filepath.Join(tempRoot, fmt.Sprintf("negative-%d", index))
		if err := copyDir(solution, negative); err != nil {
			return true, true, false, err
		}
		source, err := fixturePath(root, mutablePath)
		if err != nil {
			return true, true, false, err
		}
		target, err := fixturePath(negative, mutablePath)
		if err != nil {
			return true, true, false, err
		}
		content, err := os.ReadFile(source)
		if err != nil {
			return true, true, false, err
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return true, true, false, err
		}
		if runProof(negative, spec.Verify, spec.RequiredTests).Passed {
			return true, true, false, nil
		}
	}
	return true, true, true, nil
}

func applyQualificationSolution(workspace string, spec fixtureSpec) error {
	solutionRoot, err := filepath.Abs(spec.SolutionRoot)
	if err != nil {
		return err
	}
	for _, relative := range spec.MutablePaths {
		source, err := fixturePath(solutionRoot, relative)
		if err != nil {
			return err
		}
		target, err := fixturePath(workspace, relative)
		if err != nil {
			return err
		}
		info, err := os.Lstat(source)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("qualification solution %q must be a regular non-symlink file", relative)
		}
		resolvedRoot, err := filepath.EvalSymlinks(solutionRoot)
		if err != nil {
			return err
		}
		resolvedSource, err := filepath.EvalSymlinks(source)
		if err != nil {
			return err
		}
		relativeResolved, err := filepath.Rel(resolvedRoot, resolvedSource)
		if err != nil || relativeResolved == ".." || strings.HasPrefix(relativeResolved, ".."+string(filepath.Separator)) {
			return fmt.Errorf("qualification solution %q escapes solution root", relative)
		}
		content, err := os.ReadFile(resolvedSource)
		if err != nil {
			return fmt.Errorf("read qualification solution %q: %w", relative, err)
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func fixturePath(root, relative string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("fixture path %q escapes root", relative)
	}
	path := filepath.Join(root, clean)
	relativeToRoot, err := filepath.Rel(root, path)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("fixture path %q escapes root", relative)
	}
	return path, nil
}

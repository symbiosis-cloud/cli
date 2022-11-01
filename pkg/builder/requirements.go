package builder

import (
	"fmt"
	"os"
	"os/exec"
)

type Requirement interface {
	Value() string
	Check() bool
	Failed() error
}

type CommandRequirement struct {
	Command string
}

func (r *CommandRequirement) Value() string {
	return r.Command
}

func (r *CommandRequirement) Check() bool {
	if _, err := exec.LookPath(r.Value()); err != nil {
		return false
	}
	return true
}

func (r *CommandRequirement) Failed() error {
	return fmt.Errorf("Requirement not met. Command %s not present on this system", r.Value())
}

type FileRequirement struct {
	Path string
}

func (r *FileRequirement) Value() string {
	return r.Path
}

func (r *FileRequirement) Check() bool {
	return FileExists(r.Value())
}

func (r *FileRequirement) Failed() error {
	return fmt.Errorf("Requirement not met. File %s not found", r.Value())
}

func meetsRequirements(b Builder) error {
	requirements, err := b.requirements()

	if err != nil {
		return err
	}

	for _, requirement := range requirements {
		if !requirement.Check() {
			return requirement.Failed()
		}
	}

	return nil
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

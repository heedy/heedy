package util

import(
	"os"
	"path/filepath"
	"os/exec"
	)



// Sets the present working directory to the path of the executable
func SetWdToExecutable() error {
	path, _ := exec.LookPath(os.Args[0])
	fp, _ := filepath.Abs(path)
	dir, _ := filepath.Split(fp)
	return os.Chdir(dir)
}

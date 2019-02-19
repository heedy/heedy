package assets

import "fmt"

// WriteConfig writes the updates available in the given configuration to the given file.
// It overwrites just the updated values, leaving all others intact
func WriteConfig(filename string, c *Configuration) error {
	/*
		f, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		writer, diag := hclwrite.ParseConfig(f, filename, hcl.Pos{Line: 1, Column: 1})
		if diag != nil {
			return diag
		}
		body := writer.Body()

		// Aaaand we're fucked, because we can't write into blocks
	*/
	return fmt.Errorf("Writer not implemented")
}

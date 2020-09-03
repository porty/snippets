package templating

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/spf13/afero"
)

var funcmap = map[string]interface{}{
	"kebab": strcase.ToKebab,
	"lower": strings.ToLower,
}

// Render copies the file tree from "from" to "to" while running text/template against any ".tmpl" files found (removing the .tmpl extension afterward)
func Render(from string, to string, data interface{}) error {
	fromFS := afero.NewBasePathFs(afero.NewOsFs(), from)
	toFS := afero.NewBasePathFs(afero.NewOsFs(), to)
	return RenderFS(fromFS, toFS, data)
}

// RenderFS is the same as Render but takes afero.Fs for from/to instead of filesystem paths
func RenderFS(from afero.Fs, to afero.Fs, data interface{}) error {
	err := afero.Walk(from, "", func(fullname string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error received from file walk: %w", err)
		}
		if fullname == "" || info.Name() == "" {
			return nil
		}

		if info.IsDir() {
			if err := to.MkdirAll(fullname, info.Mode().Perm()); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".tmpl") {
			// newName := strings.TrimSuffix(info.Name(), ".tmpl")
			newPath := strings.TrimSuffix(fullname, ".tmpl")
			raw, err := afero.ReadFile(from, fullname)
			if err != nil {
				return fmt.Errorf("failed to read template %q: %w", fullname, err)
			}

			tmpl, err := template.New("file").Funcs(funcmap).Parse(string(raw))
			if err != nil {
				return fmt.Errorf("failed to parse template %q: %w", fullname, err)
			}

			// using OpenFile over Create in order to keep the file mode
			f, err := to.OpenFile(newPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return fmt.Errorf("failed to create destination file %q: %w", newPath, err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("Failed to close file %q: %s", newPath, err.Error())
				}
			}()
			if err := tmpl.Execute(f, data); err != nil {
				return fmt.Errorf("failed to execute template %q: %w", fullname, err)
			}
		} else {
			// using OpenFile over Create in order to keep the file mode
			outFile, err := to.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return fmt.Errorf("failed to create destination file %q: %w", fullname, err)
			}
			defer func() {
				if err := outFile.Close(); err != nil {
					log.Printf("Failed to close file %q: %s", fullname, err.Error())
				}
			}()
			inFile, err := from.Open(fullname)
			if err != nil {
				return fmt.Errorf("failed to open source file %q: %w", fullname, err)
			}
			defer inFile.Close()

			if _, err := io.Copy(outFile, inFile); err != nil {
				return fmt.Errorf("failed to copy file %q: %w", fullname, err)
			}
		}

		return nil
	})

	return err
}

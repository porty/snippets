package templating

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

const (
	expectedFile1 = `This file isn't templated, it shouldn't be parsed.

{{ .ThingThatDoesNotExist }}
`
	expectedFile2 = `This file *is* templated.

bonjour
`
	expectedFile3 = `This file is also templated

g'day
`
)

func TestRenderFS(t *testing.T) {
	data := map[string]interface{}{
		"Hello":  "bonjour",
		"Hello2": "g'day",
	}

	t.Run("empty", func(t *testing.T) {
		from := afero.NewMemMapFs()
		to := afero.NewMemMapFs()

		err := RenderFS(from, to, data)

		require.NoError(t, err)
		infos, err := afero.ReadDir(to, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(infos))
	})

	t.Run("not empty", func(t *testing.T) {
		from := afero.NewBasePathFs(afero.NewOsFs(), "./testdata")
		to := afero.NewMemMapFs()

		err := RenderFS(from, to, data)

		require.NoError(t, err)

		expecteds := map[string]string{
			"file1.txt":     expectedFile1,
			"file2.txt":     expectedFile2,
			"dir/file3.txt": expectedFile3,
		}

		for name, expected := range expecteds {
			b, err := afero.ReadFile(to, name)
			require.NoError(t, err, "file %q", name)
			require.Equal(t, expected, string(b))
		}
	})
}

func TestRender(t *testing.T) {
	data := map[string]interface{}{
		"Hello":  "bonjour",
		"Hello2": "g'day",
	}

	err := Render("testdata", "rofl", data)
	require.NoError(t, err)
}

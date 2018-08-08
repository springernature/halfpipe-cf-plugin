package manifest

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type ReaderWriter interface {
	ReadManifest(path string) (Manifest, error)
	WriteManifest(path string, application Application) (error)
}

type manifestReadWrite struct {
	fs afero.Afero
}

func NewManifestReadWrite(fs afero.Afero) ReaderWriter {
	return manifestReadWrite{
		fs: fs,
	}
}

func (m manifestReadWrite) ReadManifest(path string) (man Manifest, err error) {
	manifestBytes, err := m.fs.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(manifestBytes, &man)
	return
}

func (m manifestReadWrite) WriteManifest(path string, application Application) (err error) {
	manifest := Manifest{
		Applications: []Application{
			application,
		},
	}

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return
	}

	return m.fs.WriteFile(path, out, 0666)
}
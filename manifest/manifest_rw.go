package manifest

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type ReaderWriter interface {
	ReadManifest(path string) (Manifest, error)
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
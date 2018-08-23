package translatableerror

type MultipleBuildpacksFoundError struct {
	BuildpackName string
}

func (MultipleBuildpacksFoundError) Error() string {
	return "Multiple buildpacks named {{.BuildpackName}} found."
}

func (e MultipleBuildpacksFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BuildpackName": e.BuildpackName,
	})
}

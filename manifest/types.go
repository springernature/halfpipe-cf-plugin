package manifest

// These datatypes are copied from
// code.cloudfoundry.org/cli/util/manifest/raw_manifest_application.go
// As they are unexported and we cannot reuse manifest.Application because they used some weird data types that
// creates a manifest that the cf cli rejects when marshalling.

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	Name                    string             `yaml:"name,omitempty"`
	Buildpack               string             `yaml:"buildpack,omitempty"`
	Command                 string             `yaml:"command,omitempty"`
	DeprecatedDomain        interface{}        `yaml:"domain,omitempty"`
	DeprecatedDomains       interface{}        `yaml:"domains,omitempty"`
	DeprecatedHost          interface{}        `yaml:"host,omitempty"`
	DeprecatedHosts         interface{}        `yaml:"hosts,omitempty"`
	DeprecatedNoHostname    interface{}        `yaml:"no-hostname,omitempty"`
	DiskQuota               string             `yaml:"disk_quota,omitempty"`
	Docker                  DockerInfo      `yaml:"docker,omitempty"`
	EnvironmentVariables    map[string]string  `yaml:"env,omitempty"`
	HealthCheckHTTPEndpoint string             `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckType         string             `yaml:"health-check-type,omitempty"`
	Instances               int               `yaml:"instances,omitempty"`
	Memory                  string             `yaml:"memory,omitempty"`
	NoRoute                 bool               `yaml:"no-route,omitempty"`
	Path                    string             `yaml:"path,omitempty"`
	RandomRoute             bool               `yaml:"random-route,omitempty"`
	Routes                  []Route `yaml:"routes,omitempty"`
	Services                []string           `yaml:"services,omitempty"`
	StackName               string             `yaml:"stack,omitempty"`
	Timeout                 int                `yaml:"timeout,omitempty"`
}

type Route struct {
	Route string `yaml:"route"`
}

type DockerInfo struct {
	Image    string `yaml:"image,omitempty"`
	Username string `yaml:"username,omitempty"`
}


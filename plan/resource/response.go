package resource

import "time"

type Response struct {
	Version  Version
	Metadata []MetadataPair
}

type Version struct {
	Timestamp time.Time
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

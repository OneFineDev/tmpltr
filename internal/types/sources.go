package types

type SourceType string

type (
	GitSource  Source
	FileSource Source
	BlobSource Source
)

const (
	GitSourceType  SourceType = "git"
	FileSourceType SourceType = "file"
	BlobSourceType SourceType = "blob"
)

/*
Source represents the source of a set of template files that will be rendered together
*/
type Source struct {
	SourceType      `       json:"source_type"       yaml:"sourceType"`
	URL             string `json:"url"               yaml:"url"`
	Alias           string `json:"alias"             yaml:"alias"`
	Path            string `json:"path"              yaml:"path"`
	*SourceAuth     `       json:"-"                 yaml:",inline"`
	SourceAuthAlias string `json:"source_auth_alias" yaml:"sourceAuthAlias"`
}

/*
SourceAuth represents the authentication details for the source in which it is embedded
*/
type SourceAuth struct {
	AuthAlias string `json:"auth_alias"   yaml:"authAlias"`
	UserName  string `json:"username"     yaml:"userName"`
	Pat       string `json:"pat"          yaml:"pat"`
	SshKey    string `json:"ssh_key_path" yaml:"sshKeyPath"`
	Key       string `json:"key"          yaml:"key"`
	Token     string `json:"token"        yaml:"token"`
}

/*
SourceSet represents a collection of sources that collectively represent a project.
When a SourceSet in specified in a command, all Sources in that set will be fetched and rendered.
*/
type SourceSet struct {
	Alias   string            `json:"alias"   yaml:"alias"`
	Sources []string          `json:"sources" yaml:"sources"`
	Values  map[string]string `json:"values"  yaml:"values"`
}

type Sources []Source
type SourceSets []SourceSet
type SourceAuths []SourceAuth

func (t SourceConfig) Yamafiable() {}

type SourceConfig struct {
	SourceAuths SourceAuths `json:"source_auths" yaml:"sourceAuths"`
	Sources     Sources     `json:"sources"      yaml:"sources"`
	SourceSets  SourceSets  `json:"source_sets"  yaml:"sourceSets"`
}

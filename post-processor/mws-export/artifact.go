package mwsexport

import "fmt"

const BuilderId = "packer.post-processor.mws-export"

type Artifact struct {
	path string
	url  string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Id() string {
	return a.url
}

func (a *Artifact) Files() []string {
	return []string{a.path}
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Exported artifact in: %s", a.path)
}

func (*Artifact) State(name string) any {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}

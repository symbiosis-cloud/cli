package builder

import "github.com/symbiosis-cloud/cli/pkg/identity"

type Builder interface {
	Build() error
	Deploy() error
	GetIdentity() *identity.ClusterIdentity
	SetIdentity(identity *identity.ClusterIdentity)

	requirements() ([]Requirement, error)
}

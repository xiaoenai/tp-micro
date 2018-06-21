// package __TPL__ is the project template
package __TPL__

// __API__PULL__ register PULL router:
type __API__PULL__ interface {
	// IsGray check whether the service should use grayscale based on the uid.
	IsGray(*IsGrayArgs) *IsGrayResult
	// Get get the rule of gray.
	Get(*GetArgs) *GrayMatch
	// Delete delete the rule of gray.
	Delete(*DeleteArgs) *struct{}
	// Set insert or update the regular expression for matching the URI.
	Set(*SetArgs) *struct{}
}

type (
	// IsGrayArgs is_gray API parameters
	IsGrayArgs struct {
		Uri string `param:"<nozero> <rerr:400:Uri can not be empty>"`
		Uid string `param:"<nozero> <rerr:400:Uid can not be empty>"`
	}
	// IsGrayResult is_gray API result
	IsGrayResult struct {
		Gray bool
	}
	// GetArgs get API parameters
	GetArgs struct {
		Uri string `param:"<nonzero> <rerr:400:Uri can not be empty>"`
	}
	// DeleteArgs delete API parameters
	DeleteArgs struct {
		Uri string `param:"<nonzero> <rerr:400:Uri can not be empty>"`
	}
	// SetArgs set API parameters
	SetArgs struct {
		Uri    string `param:"<nonzero> <rerr:400:Uri can not be empty>"`
		Regexp string
	}
)

// MODEL create model
type __MODEL__ struct {
	GrayMatch
}

// GrayMatch
type GrayMatch struct {
	Uri    string
	Regexp string
}

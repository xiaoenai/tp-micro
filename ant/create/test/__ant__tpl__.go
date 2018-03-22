// package __ANT__TPL__ is the project template
package __ANT__TPL__

// __API__PULL__ register PULL router:
//  /home
//  /math/divide
type __API__PULL__ interface {
	Home(*struct{}) *HomeReply
	Math
}

// __API__PUSH__ register PUSH router:
//  /stat
type __API__PUSH__ interface {
	Stat(*StatArgs)
}

// Math controller
type Math interface {
	// Divide handler
	Divide(*DivideArgs) *DivideReply
}

// HomeReply home reply
type HomeReply struct {
	Content string // text
}

type (
	// DivideArgs divide api args
	DivideArgs struct {
		// dividend
		A float64
		// divisor
		B float64 `param:"<range: 0.01:100000>"`
	}
	// DivideReply divide api result
	DivideReply struct {
		// quotient
		C float64
	}
)

// StatArgs stat handler args
type StatArgs struct {
	Ts int64 // timestamps
}

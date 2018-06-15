// package __ANT__TPL__ is the project template
package __ANT__TPL__

// __API__PULL__ register PULL router:
//  /home
//  /math/divide
type __API__PULL__ interface {
	Home(*struct{}) *HomeResult
	Math
}

// __API__PUSH__ register PUSH router:
//  /stat
type __API__PUSH__ interface {
	Stat(*StatArg)
}

// MODEL create model
type __MODEL__ struct {
	DivideArg
	User
}

// Math controller
type Math interface {
	// Divide handler
	Divide(*DivideArg) *DivideResult
}

// HomeResult home result
type HomeResult struct {
	Content string // text
}

type (
	// DivideArg divide api arg
	DivideArg struct {
		// dividend
		A float64
		// divisor
		B float64 `param:"<range: 0.01:100000>"`
	}
	// DivideResult divide api result
	DivideResult struct {
		// quotient
		C float64
	}
)

// StatArg stat handler arg
type StatArg struct {
	Ts int64 // timestamps
}

// User user info
type User struct {
	Id   int64
	Name string
	Age  int32
}

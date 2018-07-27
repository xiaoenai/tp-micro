// Command project is the tp-micro service project.
// The framework reference: https://github.com/xiaoenai/tp-micro
package __TPL__

// __API_CALL__ register CALL router
type __API_CALL__ interface {
	// Home handler
	Home(*struct{}) *HomeResult
	// Math controller
	Math
}

// __API_PUSH__ register PUSH router:
//  /stat
type __API_PUSH__ interface {
	Stat(*StatArg)
}

// __MYSQL_MODEL__ create mysql model
type __MYSQL_MODEL__ struct {
	User
	Log
	Device
}

// __MONGO_MODEL__ create mongodb model
type __MONGO_MODEL__ struct {
	Meta
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
	Ts int64 `param:"<query:ts>"` // timestamps
}

// User user info
type User struct {
	Id   int64  `key:"pri"`
	Name string `key:"uni"`
	Age  int32
}

type Log struct {
	Text string
}

type Device struct {
	UUID string `key:"pri"`
}

type Meta struct {
	Hobby []string
	Tags  []string
}

package types

type (
	// MathDivideArgs args
	MathDivideArgs struct {
		A int
		B int `param:"<range:1:>"`
	}
	// MathDivideReply reply
	MathDivideReply = int
)

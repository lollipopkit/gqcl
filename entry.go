package gqcl

// Evaluate parses and evaluates an expression against the provided context Value.
func Evaluate(expression string, ctx Value) (Value, error) {
	expr, err := Parse(expression)
	if err != nil {
		return Value{}, err
	}
	return expr.Eval(ctx)
}

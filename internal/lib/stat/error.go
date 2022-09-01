package stat

type ParseError struct {
	msg string
}

var _ error = &ParseError{}

func newParseError(msg string) *ParseError {
	return &ParseError{msg: msg}
}

func (p *ParseError) Error() string {
	s := packageName + "stat.ParseError"
	if p.msg != "" {
		s = s + ": " + p.msg
	}
	return s
}

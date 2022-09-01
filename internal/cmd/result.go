package cmd

type Result interface {
	ExitCode() int
}

type result struct {
	exitCode int
}

func NewResult(exitCode int) Result {
	return &result{
		exitCode: exitCode,
	}
}

func (c *result) ExitCode() int {
	return c.exitCode
}

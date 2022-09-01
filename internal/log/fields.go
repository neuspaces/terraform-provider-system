package log

type Field func(f map[string]interface{})

func Package(pkg string) Field {
	return func(f map[string]interface{}) {
		f["package"] = pkg
	}
}

func Error(err error) Field {
	return func(f map[string]interface{}) {
		f["error"] = err.Error()
	}
}

// AdditionalFields creates a map[string]interface{} to be passed to a log function of the tflog package
func AdditionalFields(fields ...Field) map[string]interface{} {
	fm := map[string]interface{}{}
	for _, field := range fields {
		field(fm)
	}
	return fm
}

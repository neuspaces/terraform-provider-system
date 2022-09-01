package heredoc

// String returns un-indented string as here-document.
// String is an alias of Doc
func String(raw string) string {
	return Doc(raw)
}

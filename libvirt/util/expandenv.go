package util

// ExpandEnvExt expands environment variables and resolves ~ to the home directory
// this is a drop-in replacement for os.ExpandEnv but is additionall '~' aware
func ExpandEnvExt(path string) string {
	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // return path as-is if unable to resolve home directory
		}
		// Replace ~ with home directory
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}

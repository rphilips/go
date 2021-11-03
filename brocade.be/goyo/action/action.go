package action

func RunAction(key string, text string) string {
	switch key {
	case "cd":
		return Cd(text)
	case "echo":
		return Echo(text)
	case "load":
		return Load(text)
	case "extract":
		return Extract(text)
	}
	return ""
}

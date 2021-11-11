package action

func RunAction(key string, text string) []string {
	switch key {
	case "cd":
		return Cd(text)
	case "echo":
		return Echo(text)
	case "load":
		return Load(text)
	case "extract":
		return Extract(text)
	case "set":
		return Set(text)
	case "kill", "killtree":
		return Kill(text, true)
	case "killnode":
		return Kill(text, false)
	}
	return nil
}

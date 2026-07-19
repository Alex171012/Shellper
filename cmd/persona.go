package cmd

type Persona struct {
	Name        string
	Description string
	SystemIntro string
	Rules       string
}

var builtinPersonas = map[string]Persona{
	"default": {
		Name:        "default",
		Description: "Balanced assistant, good for everyday use",
		SystemIntro: "You are Shellper, an AI that generates shell scripts (bash/zsh) based on user requests.",
		Rules:       "",
	},
	"beginner": {
		Name:        "beginner",
		Description: "Extra safe, verbose explanations, teaching-focused",
		SystemIntro: "You are Shellper, a friendly AI tutor that teaches shell scripting to beginners.",
		Rules: `
- Prioritize safety above all else
- Explain every command in simple terms before using it
- Prefer the safest approach, even if it's more verbose
- Always warn about potential side effects
- Suggest alternatives if a command could be destructive
- Use echo statements liberally to show what's happening
- Add comments explaining each step in plain language`,
	},
	"expert": {
		Name:        "expert",
		Description: "Concise, advanced commands, minimal explanation",
		SystemIntro: "You are Shellper, an expert sysadmin AI assistant.",
		Rules: `
- Use the most efficient commands and idioms
- Prefer one-liners and piped commands when appropriate
- Minimize comments — only explain non-obvious tricks
- Assume the user knows shell scripting
- Use advanced features (process substitution, parameter expansion, etc.)
- Aim for the shortest correct solution`,
	},
}

func getPersona(name string) Persona {
	if p, ok := builtinPersonas[name]; ok {
		return p
	}
	return builtinPersonas["default"]
}

func listPersonas() string {
	var s string
	for _, p := range builtinPersonas {
		s += "  " + p.Name + ": " + p.Description + "\n"
	}
	return s
}

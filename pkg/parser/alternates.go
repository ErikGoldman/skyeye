package parser

// alternateRequestWords is a map of alternate forms of request words.
// These are used to provide aliases for certain commands and to deal with quality issues in speech-to-text.
var alternateRequestWords = map[string]string{
	"alphacheck":         alphaCheck,
	"bog it":             bogeyDope,
	"bogado":             bogeyDope,
	"bogeido":            bogeyDope,
	"bogeied":            bogeyDope,
	"bogey-doke":         bogeyDope,
	"bogeydoke":          bogeyDope,
	"bogeydope":          bogeyDope,
	"bogeydote":          bogeyDope,
	"bogeyedope":         bogeyDope,
	"boggy":              bogeyDope,
	"bogido":             bogeyDope,
	"bogie":              bogeyDope,
	"bogit":              bogeyDope,
	"bogota":             bogeyDope,
	"bogueed":            bogeyDope,
	"bogueto":            bogeyDope,
	"bogy":               bogeyDope,
	"boji":               bogeyDope,
	"bokeh":              bogeyDope,
	"bokeido":            bogeyDope,
	"bokey":              bogeyDope,
	"bokeydope":          bogeyDope,
	"booby dop":          bogeyDope,
	"boog it":            bogeyDope,
	"boogie":             bogeyDope,
	"book it out":        bogeyDope,
	"book it up":         bogeyDope,
	"bovido":             bogeyDope,
	"boy dope":           bogeyDope,
	"bubby dope":         bogeyDope,
	"bucket up":          bogeyDope,
	"bug it out":         bogeyDope,
	"bug it":             bogeyDope,
	"bugadope":           bogeyDope,
	"bugged up":          bogeyDope,
	"buggettope":         bogeyDope,
	"buggidop":           bogeyDope,
	"buggie do":          bogeyDope,
	"buggie dog":         bogeyDope,
	"buggie dope":        bogeyDope,
	"buggy do":           bogeyDope,
	"buggy dog":          bogeyDope,
	"buggy dope":         bogeyDope,
	"checking in":        checkIn,
	"com check":          radioCheck,
	"comcheck":           radioCheck,
	"comes check":        radioCheck,
	"comm":               radioCheck,
	"comms":              radioCheck,
	"commscheck":         radioCheck,
	"commshack":          radioCheck,
	"comp check":         radioCheck,
	"comps check":        radioCheck,
	"coms":               radioCheck,
	"comsjack":           radioCheck,
	"declared":           declare,
	"fogey":              bogeyDope,
	"fogeyed":            bogeyDope,
	"foggy":              bogeyDope,
	"foggydope":          bogeyDope,
	"fogy dope":          bogeyDope,
	"fogy":               bogeyDope,
	"lucky dope":         bogeyDope,
	"mogito":             bogeyDope,
	"ogi doke":           bogeyDope,
	"ogi dop":            bogeyDope,
	"ogi dope":           bogeyDope,
	"ogidope":            bogeyDope,
	"okey":               bogeyDope,
	"oogie":              bogeyDope,
	"perimeter":          tripwire,
	"pogito":             bogeyDope,
	"pogy dope":          bogeyDope,
	"pokedo":             bogeyDope,
	"pokedome":           bogeyDope,
	"radiocheck":         radioCheck,
	"read a check":       radioCheck,
	"read it check":      radioCheck,
	"set threat radius":  tripwire,
	"set warning radius": tripwire,
	"set warning":        tripwire,
	"snap lock":          snaplock,
	"trip wire":          tripwire,
	"voki":               bogeyDope,
	"warn me":            tripwire,
}

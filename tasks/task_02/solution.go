package main

func rotateRunes(s string, shift int) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	shift = ((shift % len(runes)) + len(runes)) % len(runes) // чтобы избежать проблем с MinInt и сдвиг вправо(отр.)
	if shift != 0 {
		return string(runes[shift:]) + string(runes[:shift])
	}

	return string(runes)

}

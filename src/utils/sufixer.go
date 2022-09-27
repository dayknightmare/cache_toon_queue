package utils

func Sufixer(name string, typeQueue string, priority int) string {
	sufix := "f"

	if typeQueue == "lifo" {
		sufix = "l"
	}

	if priority >= 1 {
		sufix = "z" + sufix
	}

	return name + ":" + sufix
}

package namer

import (
	"fmt"
	"math/rand"
)

func Namer() {
	rand.Seed(42)
	names := []string{
		"sds",
	}
	fmt.Println(names[rand.Intn(len(names))])
}

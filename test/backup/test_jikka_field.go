package main

import "dis-core/internal/jikka"

func main() {
	jikka.CreateJikka(
		"domain.human.usa",
		"domain.human.europa",
		true, // freeWillA
		true, // freeWillB
		true, // recognition
		true, // consent
		true, // reflection
	)
}

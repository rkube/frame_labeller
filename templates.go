package main

import (
	"os"
	"text/template"
)

type state_data_2 struct {
	Username string
	Id       int
}

func main() {
	// dogs := []Pet{
	// 	{
	// 		Name:   "Jujube",
	// 		Sex:    "Female",
	// 		Intact: false,
	// 		Age:    "10 months",
	// 		Breed:  "German Shepherd/Pitbull",
	// 	},
	// 	{
	// 		Name:   "Zephyr",
	// 		Sex:    "Male",
	// 		Intact: true,
	// 		Age:    "13 years, 3 months",
	// 		Breed:  "German Shepherd/Border Collie",
	// 	},
	// }

	my_state := state_data_2{Username: "anonymous", Id: 0}

	var tmplFile = "pets.tmpl"
	tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, my_state)
	if err != nil {
		panic(err)
	}
} // end main

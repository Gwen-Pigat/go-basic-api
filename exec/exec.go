package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {

	mod := strings.ToLower(os.Args[1])
	if mod == "" {
		log.Fatal("Vous devez spécifier un argument")
	}
	repMod := "modules/" + mod
	err := os.Mkdir(repMod, 0777)
	if err != nil {
		log.Fatalf("Rep %v cannot be created", mod)
	}
	content, err := os.ReadFile("module.txt")
	if err != nil {
		log.Fatal("Base file cannot be opened")
	}
	file, err := os.Create(repMod + "/" + mod + ".go")
	if err != nil {
		log.Fatalf("File %v cannot be created", file.Name())
	}
	caser := cases.Title(language.English)
	contentStr := string(content)
	contentStr = strings.ReplaceAll(contentStr, "nameSmall", mod)
	contentStr = strings.ReplaceAll(contentStr, "NameUp", caser.String(mod))
	_, err = file.Write([]byte(contentStr))
	if err != nil {
		log.Fatal("new file cannot be set")
	}
	fmt.Printf("%v crée", mod)
}

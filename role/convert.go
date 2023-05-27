package role

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func Convert(src, dst string) {
	// Open the file for reading
	file, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Create a slice to hold the roles
	var roles = make(Roles)
	var role = &Role{}
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			if role.Name != "" && role.Desc != "" {
				roles[role.Name] = role.Desc
			}
			role.Name = ""
			role.Desc = ""
			continue
		}

		if role.Name == "" {
			role.Name = strings.Trim(scanner.Text(), "\n")
		} else {
			desc := strings.Trim(scanner.Text(), "\r\n")
			role.Desc += strings.Trim(desc, "\n")
		}
		fmt.Println("...........", role.Name)
		fmt.Println("...........", role.Desc)
	}
	// Marshal the slice into YAML format
	yamlContent, err := yaml.Marshal(&roles)
	if err != nil {
		log.Fatal(err)
	}

	// Write the YAML content to a file
	err = os.WriteFile(dst, yamlContent, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File converted to YAML format successfully!")
}

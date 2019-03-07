package main

import (
	"os"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"bitbucket.org/araneaws/margono"
	"gopkg.in/urfave/cli.v1"
)

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func exist(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// ArtefactPath Get full path of artefact
func ArtefactPath() string {
	cnf := libs.GetConfig()
	loc := filepath.Join(cnf.Main.Artefact, "users")

	return loc
}

// Push new account to database
func Push(rec *libs.Account) error {

	recj, err := json.Marshal(rec)

	if !exist(ArtefactPath()) {
		os.MkdirAll(ArtefactPath(), 0700)
	}
	err = ioutil.WriteFile(filepath.Join(ArtefactPath(), rec.Os.Name), recj, 0600)

	return err
}

// Pull account from database
func Pull(name string) libs.Account {
	var a libs.Account
	//	var out interface{}

	rawjson, err := ioutil.ReadFile(filepath.Join(ArtefactPath(), name))
	check(err)

	err = json.Unmarshal(rawjson, &a)
	//	fmt.Printf("%+v", a.Domain)
	//	fmt.Printf("%+v", err)
	check(err)

	//	a.Domain = out["Domain"]
	//	a = out.(libs.Account)

	//  if require a list, this is the method.
	//
	//	if strings.Compare(name, "all") {
	//		// baca setiap file dan masukkan hasilnya ke array
	//	} else if list {
	//		// baca list file-nya saja.
	//		if strings.Compare(name, "all") {
	//			// baca list seluruh file
	//			users, err := ioutil.ReadDir(ArtefactPath())
	//			check(err)
	//			for _, user := range users {
	//				append(out, user.Name())
	//			}
	//			return out
	//		} else {
	//			// baca list satu file saja
	//		}
	//	} else {
	//		// baca satu file saja
	//	}

	return a
}

// down the service
func down(rec *libs.Account) error {

}

func main() {
	susy := cli.NewApp()
	susy.Name = "susy"
	susy.Usage = "Aranea site user system"
	susy.Version = "0.1.0"

	susy.Commands = []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create new user to the system.",
			Action: func(c *cli.Context) error {
				var out error
				a := libs.New(c.Args().First())
				_, out = a.CreateUserOs()
				out = a.CreateUserDb()
				check(out)
				out = a.DeployAsset()
				check(out)

				af, _ := json.MarshalIndent(a, "", "  ")
				fmt.Println(string(af))
				Push(a)
				return out
			},
		},
		{
			Name:    "read",
			Aliases: []string{"r"},
			Usage:   "Read user from the system.",
			Action: func(c *cli.Context) error {
				var a libs.Account
				in := c.Args().First()
				if strings.Index(in, ".") > -1 {
					acc := libs.New(in)
					a = Pull(acc.Os.Name)
				} else {
					a = Pull(in)
				}

				out, err := json.MarshalIndent(a, "", "  ")
				check(err)

				fmt.Println(string(out))

				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l", "ls"},
			Usage:   "susy list",
			Action: func(c *cli.Context) error {
				files, err := ioutil.ReadDir(ArtefactPath())
				check(err)
				for _, file := range files {
					fmt.Println(file.Name())
				}

				return nil
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Remove user from the system.",
			Action: func(c *cli.Context) error {
				a := libs.New(c.Args().First())
				a.RemoveUserOs()
				a.RemoveUserDb()
				err := os.Remove(filepath.Join(ArtefactPath(), a.Os.Name))
				check(err)
				fmt.Println("User deleted", a.Domain)
				//				out, _ := json.MarshalIndent(a, "", "  ")
				//				fmt.Println(string(out))

				return nil
			},
		},
		{
			Name:    "deleteall",
			Aliases: []string{"da", "dd"},
			Usage:   "susy da",
			Action: func(c *cli.Context) error {
				files, err := ioutil.ReadDir(ArtefactPath())
				check(err)
				for _, file := range files {

					a := libs.New(file.Name())
					a.RemoveUserOs()
					a.RemoveUserDb()
					err := os.Remove(filepath.Join(ArtefactPath(), a.Os.Name))
					check(err)
					fmt.Println("User deleted", a.Domain)
					//				out, _ := json.MarshalIndent(a, "", "  ")
					//				fmt.Println(string(out))
				}

				return nil
			},
		},
	}

	susy.Run(os.Args)
}

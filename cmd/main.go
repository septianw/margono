package main

import (
	"os"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	//	"syscall"

	//	"golang.org/x/net/context"

	//	"github.com/docker/libcompose/docker"
	//	"github.com/docker/libcompose/docker/ctx"
	//	"github.com/docker/libcompose/project"
	//	"github.com/docker/libcompose/project/options"

	"bitbucket.org/araneaws/margono"
	"github.com/dustinkirkland/golang-petname"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

var (
	webports []uint16
	sshport  uint16
	debug    bool = false
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

func ArtefactPath() string {
	cnf := libs.GetConfig()
	loc := filepath.Join(cnf.Main.Artefact, "users")

	return loc
}

func Push(rec *libs.Account) error {

	recj, err := json.Marshal(rec)
	//	fmt.Println("push")

	if !exist(ArtefactPath()) {
		os.MkdirAll(ArtefactPath(), 0700)
	}
	err = ioutil.WriteFile(filepath.Join(ArtefactPath(), rec.Os.Name), recj, 0600)

	return err
}

func Pull(name string) (*libs.Account, error) {
	//	fmt.Println("pull")

	var a libs.Account
	//	var out interface{}

	rawjson, err := ioutil.ReadFile(filepath.Join(ArtefactPath(), name))
	check(err)

	err = json.Unmarshal(rawjson, &a)
	//	fmt.Printf("%+v", a.Domain)
	//	fmt.Printf("%+v", err)
	//	check(err)

	return &a, err
}

func GetName() string {
	return petname.Generate(3, "-")
}

func GetDomainName(domain string) string {
	if strings.Compare(domain, "") == 0 {
		domain = GetName() + "." + libs.GetConfig().Main.Domain
	}

	return domain
}

func GetAccount(domain string) (*libs.Account, error) {
	var err error

	if strings.Compare(domain, "") == 0 {
		err = errors.New("Domain empty, GetAccount require a domain name")
		return nil, err
	}

	var account = libs.New(domain)

	if (strings.Compare(account.Os.Name, "") != 0) && (exist(path.Join(ArtefactPath(), account.Os.Name))) {
		account, err = Pull(account.Os.Name)
		check(err)
	} else {
		_, err = account.CreateUserOs()
		check(err)
	}

	return account, err
}

func GenApp(app App, acc *libs.Account) Web {
	var web Web
	var err error

	var homeDir = acc.Os.Home
	var env = map[string]string{
		"DBUSER": acc.Db.User,
		"DBPASS": acc.Db.Pass,
		"DBHOST": acc.Db.Host,
		"DBNAME": acc.Db.Name,
		"DBPORT": strconv.Itoa(int(acc.Db.Port)),
	}

	var stor = make([]string, len(app.Storage))
	web = Web{
		Restart:       "always",
		Mem_limit:     "128M",
		Memswap_limit: "128M",
		Cpu_shares:    64,
		Cpu_quota:     25000,
	}

	for i, volume := range app.Storage {
		s := path.Join(homeDir, "space", volume)
		err = os.MkdirAll(s, 0770)
		err = os.Chown(s, acc.Os.UID, acc.Os.UID)
		err = os.Chown(path.Join(homeDir, "space"), acc.Os.UID, acc.Os.UID)
		check(err)
		stor[i] = fmt.Sprintf("%s:%s", s, path.Join("/var/www/html/", volume))
	}
	//web.Volumes = []string{}
	web.Volumes = stor
	web.Environment = env

	//	bit, _ := yaml.Marshal(webs)

	//fmt.Printf("%+v", sj.Apps[app])
	//	fmt.Printf("%+v", string(bit))

	return web
}

func ChownR(res string, uid, gid int) error {
	var out error
	out = filepath.Walk(res, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}

		return err
	})
	return out
}

// if src without trailling slash, make sure dst too
func MoveR(src, dst string) error {
	var out error
	out = filepath.Walk(src, func(name string, info os.FileInfo, err error) error {
		if strings.Compare(name, src) != 0 {
			d := strings.Replace(name, src, dst, -1)
			err = os.Rename(name, d)
		}
		return err
	})
	return out
}

// if src without trailling slash, make sure dst too
func CopyR(src, dst string) error {
	var out error
	out = filepath.Walk(src, func(name string, info os.FileInfo, err error) error {
		if strings.Compare(name, src) != 0 {
			d := strings.Replace(name, src, dst, -1)
			n, err := os.Stat(name)
			_, err = os.Stat(dst)
			if os.IsNotExist(err) {
				err = os.MkdirAll(dst, 0775)
				check(err)
			}
			if n.IsDir() { // if directory, make it in destination
				err = os.MkdirAll(d, 0775)
				check(err)
			} else {
				f, err := os.Open(name) // open source file
				check(err)
				defer f.Close()

				df, err := os.Create(d) // create destination file if not exist
				check(err)
				defer df.Close()

				fw, err := io.CopyN(df, f, n.Size()) // copy operation here
				if (fw != n.Size()) && (err != nil) {
					check(err)
				}
			}
		}
		return err
	})
	return out
}

func GetPlatformName(Platform, Type string) (string, string) {
	var version string
	switch Platform {
	case "wordpress":
		Platform = "wordpress"
		version = "latest"
	case "wordpress46":
		Platform = "wordpress"
		version = "4.6"
	case "wordpress45":
		Platform = "wordpress"
		version = "4.5"
	case "wordpress44":
		Platform = "wordpress"
		version = "4.4"
	case "wordpress43":
		Platform = "wordpress"
		version = "4.3"
	case "wordpress42":
		Platform = "wordpress"
		version = "4.2"
	case "wordpress39":
		Platform = "wordpress"
		version = "3.9"
	case "custom":
		switch Type {
		case "php":
			Platform = "apachebase"
			version = "php"
		case "php56":
			Platform = "apachebase"
			version = "php5.6"
		case "php55":
			Platform = "apachebase"
			version = "php5.5"
		}
	}

	return Platform, version
}

// ini harusnya looping isi surat jalan untuk di-append ke docker-compose
/*
	- [ ] looping isi apps untuk dimasukkan ke docker-compose.yml
	- [ ] apps1 menjadi leader/interface ke load balancer
	- [ ] jumlah isi apps ditentukan dari lumbung
	- [ ] jumlah memory per apps ditentukan dari lumbung
*/
func MakeDockerCompose(SuratJalan SuratJalan) (*libs.Account, DockerCompose) {

	// Initiate default value
	var (
		out         error
		DCompose    DockerCompose
		DComposeSsh Ssh
		DComposeLb  LoadBalancer
		dcspath     string
		uid         int
		gid         int
	)

	// set domain
	if strings.Compare(SuratJalan.Domain, "") == 0 {
		SuratJalan.Domain = petname.Generate(3, "") + "." + libs.GetConfig().Main.Domain
	}

	// if account present get it, if not create it.
	acc, out := GetAccount(SuratJalan.Domain)
	check(out)

	userOs, out := user.Lookup(acc.Os.Name)
	check(out)

	if len(SuratJalan.Dbs) != 0 {
		out = acc.CreateUserDb()
		check(out)
	}

	DComposeSsh = Ssh{
		Build: "ssh",
		//		Ports: []string{},
		//		Volumes:   []string{"/sites/tasmodelternet/space:/home/tasmodelternet"}, // FIXME ambil path dari surat jalan
		Mem_limit: "25M",
		//		Restart:   "",
	}

	DComposeLb = LoadBalancer{
		Image:       "dockercloud/haproxy:1.4",
		Restart:     "always",
		Volumes:     []string{"/var/run/docker.sock:/var/run/docker.sock"},
		Environment: []string{"STATS_AUTH=asep:ganteng"},
		//		Ports:       []string{"32788:80", "32789:1936"}, // FIXME ambil port dari susy
	}

	DCompose = DockerCompose{
		Version: "2",
	}

	webports = []uint16{
		libs.GetPort("web"),
		libs.GetPort("web"),
	}
	fmt.Printf("webport=%d\n", webports[0])
	fmt.Printf("haport=%d\n", webports[1])

	// generate ssh port if not present
	if acc.Ssh.Port == 0 {
		sshport = libs.GetPort("ssh")
		acc.Ssh.Port = sshport
	} else {
		sshport = acc.Ssh.Port
	}

	// compile components
	DComposeSsh.Ports = append(DComposeSsh.Ports, strconv.Itoa(int(sshport))+":22")
	DComposeLb.Ports = append(DComposeLb.Ports, strconv.Itoa(int(webports[0]))+":80")
	DComposeLb.Ports = append(DComposeLb.Ports, strconv.Itoa(int(webports[1]))+":1936")

	DCompose.Services = make(map[string]interface{})

	i := 0
	for app := range SuratJalan.Apps {
		a := GenApp(SuratJalan.Apps[app], acc)
		a.Build = app
		DCompose.Services[app] = a
		//		DComposeSsh.Volumes = a.Volumes
		DComposeSsh.Volumes = []string{
			fmt.Sprintf("%s:%s", path.Join(acc.Os.Home, "space"), acc.Os.Home),
		}
		os.MkdirAll(path.Join(acc.Os.Home, app), 0770)

		dcspath = path.Join(acc.Os.Home, app, "Dockerfile")

		platform, version := GetPlatformName(SuratJalan.Apps[app].Platform, SuratJalan.Apps[app].Type)

		out = ioutil.WriteFile(dcspath, []byte(acc.GetWebDockerfile(platform, version, SuratJalan.Apps[app].Deploy, SuratJalan.Apps[app].Postdeploy)), 0755)
		check(os.Rename(path.Join(acc.Os.Home, "entrypoint.sh"), path.Join(acc.Os.Home, app, "entrypoint.sh")))
		check(os.Rename(path.Join(acc.Os.Home, "postdeploy.sh"), path.Join(acc.Os.Home, app, "postdeploy.sh")))

		uid, _ = strconv.Atoi(userOs.Uid)
		gid, _ = strconv.Atoi(userOs.Gid)
		//		os.Chown(dcspath, uid, gid)
		//		check(out)
		if i == 0 {
			DComposeLb.Links = []string{app}
			cwd, _ := os.Getwd()
			out = CopyR(cwd, path.Join(acc.Os.Home, app, "artefact"))
			check(out)
		}
		out = ChownR(acc.Os.Home, uid, gid)
		check(out)
		i++
	}
	DCompose.Services["lb"] = DComposeLb
	DCompose.Services["ssh"] = DComposeSsh

	bit, out := yaml.Marshal(DCompose)

	// write docker-compose to userpath
	dcspath = path.Join(acc.Os.Home, "docker-compose.yml")
	out = ioutil.WriteFile(dcspath, bit, 0755)
	check(out)
	uid, _ = strconv.Atoi(userOs.Uid)
	gid, _ = strconv.Atoi(userOs.Gid)
	os.Chown(dcspath, uid, gid)
	check(out)

	// deploy Dockerfiles
	out = acc.DeployAsset()
	check(out)

	dcspath = path.Join(acc.Os.Home, "ssh", "Dockerfile")
	out = ioutil.WriteFile(dcspath, []byte(acc.GetSshDockerfile()), 0755)
	check(out)
	uid, _ = strconv.Atoi(userOs.Uid)
	gid, _ = strconv.Atoi(userOs.Gid)
	os.Chown(dcspath, uid, gid)
	check(out)

	check(Push(acc))

	//	fmt.Println(string(bit))

	return acc, DCompose
}

func Start(acc *libs.Account) error {
	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer

	// get account

	//	p, err := docker.NewProject(&ctx.Context{
	//		Context: project.Context{
	//			ComposeFiles: []string{path.Join(acc.Os.Home, "docker-compose.yml")},
	//			ProjectName:  acc.Domain,
	//		},
	//	}, nil)
	//	check(err)

	//	return p.Up(context.Background(), options.Up{})

	//	userOs, err := user.Lookup(acc.Os.Name)
	//	check(err)
	//	uid, _ := strconv.Atoi(userOs.Uid)
	//	gid, _ := strconv.Atoi(userOs.Gid)

	// exec docker-compose as user in account
	//	cmd := exec.Command("/bin/bash", "-x", "/usr/local/sbin/margono-start", acc.Os.Home)
	os.Chdir(acc.Os.Home)
	cmd := exec.Command("ls", "-alh")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	//	cmd.SysProcAttr = &syscall.SysProcAttr{}
	//	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}

	// return string output if can
	fmt.Printf("out: %s\n", out.String())
	fmt.Printf("err: %s\n", stderr.String())
	err = cmd.Run()

	return err
}

func main() {
	margono := cli.NewApp()
	margono.Name = "margono"
	margono.Usage = "Margono surat jalan checker"
	margono.Version = "0.1.1"

	if strings.Compare(os.Getenv("APDEBUG"), "") == 0 {
		debug = false
	} else {
		debug = true
	}

	margono.Commands = []cli.Command{
		{
			Name:    "read",
			Aliases: []string{"r"},
			Usage:   "Read surat jalan and generate new artefact files.",
			Action: func(c *cli.Context) error {
				var (
					out        error
					file       = c.Args().First()
					SuratJalan SuratJalan
				)
				c.FlagNames()
				basepath, _ := os.Getwd()
				absfile := path.Join(basepath, file)
				rawyml, _ := ioutil.ReadFile(absfile)

				out = yaml.Unmarshal(rawyml, &SuratJalan)

				if debug {
					log.Println("Processing: ", absfile)
				}

				acc, _ := MakeDockerCompose(SuratJalan)
				fmt.Printf("domain=%s\n", acc.Domain)
				fmt.Printf("name=%s\n", acc.Os.Name)

				return out
			},
		},
		{
			Name:    "run",
			Aliases: []string{"rr"},
			Usage:   "Read surat jalan, generate new artefact files, and run it.",
			Action: func(c *cli.Context) error {
				var (
					out        error
					file       = c.Args().First()
					SuratJalan SuratJalan
				)
				basepath, _ := os.Getwd()
				absfile := path.Join(basepath, file)
				rawyml, _ := ioutil.ReadFile(absfile)

				out = yaml.Unmarshal(rawyml, &SuratJalan)

				if debug {
					log.Println("Processing: ", absfile)
				}

				acc, _ := MakeDockerCompose(SuratJalan)

				//				fmt.Println(acc)
				out = Start(acc)

				return out
			},
		},
		{
			Name:    "rin",
			Aliases: []string{"ri"},
			Usage:   "Read surat jalan, generate new artefact files, and run it.",
			Action: func(c *cli.Context) error {

				if strings.Compare(os.Getenv("MARGONO_DEBUG"), "") == 0 {
					fmt.Println("0")
				} else {
					fmt.Println("1")
				}

				return nil
			},
		},
	}

	margono.Run(os.Args)
}

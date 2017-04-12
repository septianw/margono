/*
	TODOLIST:
	  - [x] Generate name
	  - [x] Generate password db
	  - [x] Generate password sftp
	  - [x] Generate password general
	  - [x] Get last UID
	  - [x] Create user OS group docker
	  - [x] set quota
	  - [x] Create user db
	  - [x] Construct directory structure
	  - [x] Write main dockerfile	// need more parameter from user
	  - [x] Write ssh dockerfile
	  - [ ] Generate docker-compose.yml file	// need more parameter from user
	  - [x] Get port from range
*/
package libs

import (
	//	"flag"
	"database/sql"
	"fmt"
	//	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	//	"testing"
	"log"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/BurntSushi/toml"
)

/*
	e		error object
	code	error code

	1 error file
	2 error db
	3 error algorithm
	4 error ssh
	5 error user
	6 error command
	9 unknown error
*/

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.,"
	NAMELEN       = 14
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits

	ERRFILE = 1
	ERRDB   = 2
	ERRALG  = 3
	ERRSSH  = 4
	ERRUSR  = 5
	ERRCMD  = 6
	ERRUNKN = 9
)

type CnfUser struct {
	Home string
}

type CnfMain struct {
	Domain   string
	Artefact string
	Dsn      string
	Dbhost   string
	Dbport   uint16
	Restart  string
	Assets   []string
}

type CnfRes struct {
	Memlimit     string
	MemswapLimit string
	Cpushares    uint8
	Cpuquota     uint16
	Sshmemlimit  string
	Statauth     string
}

type CnfPorts struct {
	Ssh string
	Web string
}

type Config struct {
	Main      CnfMain
	User      CnfUser
	Resources map[string]CnfRes
	PortRange CnfPorts
}

type AccountDb struct {
	Name string
	Host string
	Port uint16
	User string
	Pass string
}

type AccountOs struct {
	Name string
	Pass string
	UID  int
	Home string
}

type AccountSsh struct {
	Host string
	Port uint16
	User string
	Pass string
}

type Account struct {
	Domain string
	Db     AccountDb
	Os     AccountOs
	Ssh    AccountSsh
}

var debug bool = false

/*
	e		error object
	code	error code

	1 error file
	2 error db
	3 error algorithm
	4 error ssh
	5 error user
	6 error command
	9 unknown error
*/
// Done
func check(e error, code int) bool {
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(code)
		return false
	}

	return true
}

// Done
/*
	Find available port by try to listen on it.
*/
func AvailablePort(in uint16) bool {
	var out bool

	ln, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(int(in))))

	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			out = false
		} else {
			out = false // experimental
		}
	} else {
		ln.Close()
		out = true
	}

	return out
}

// Done
/*
	Check available port one by one
*/
func FindPort(in string) uint16 {
	var port uint16

	ports := strings.Split(in, "-")
	mi, _ := strconv.Atoi(ports[0])
	ma, _ := strconv.Atoi(ports[1])

	portmin := uint16(mi)
	portmax := uint16(ma)

	i := portmin

	for {

		if AvailablePort(i) {
			port = i
			break
		}

		if i == portmax {
			break
		}
		i++
	}

	return port
}

// Done
/*
	Get confing from file
*/
func GetConfig() Config {
	var config Config
	if _, e := toml.DecodeFile("/etc/margono.conf", &config); e != nil {
		check(e, ERRFILE)
	}
	return config
}

/* Done
Mendapatkan port ssh terakhir dari catatan.
*/
func GetPort(component string) uint16 {
	var cnf = GetConfig()
	var err error
	var o uint16
	var Range string
	var cnfrange string
	var fileport string

	// default range cnf.PortRange.Ssh
	// tentukan range, kalau last port kosong pakai default dari konfig
	// kalau last port ada isinya, gunakan nilai min dari last port dan max dari config.
	// port := FindPort(range)
	// saat ketemu portnya, tulis port terakhir yang digunakan ke last port.

	switch component {
	case "web":
		cnfrange = cnf.PortRange.Web
		fileport = "weblastport"
	case "ssh":
		cnfrange = cnf.PortRange.Ssh
		fileport = "sshlastport"
	default:
		return 0
	}

	lstport, err := ioutil.ReadFile(path.Join(cnf.Main.Artefact, fileport))
	l, stre := strconv.Atoi(string(lstport))

	if (err != nil) || (l == 0) || (stre != nil) {
		Range = cnfrange
	} else {
		defrange := strings.Split(cnfrange, "-")
		l = l + 1
		Range = fmt.Sprintf("%s-%s", strconv.Itoa(l), defrange[1])
	}

	o = FindPort(Range)
	err = ioutil.WriteFile(path.Join(cnf.Main.Artefact, fileport), []byte(strconv.Itoa(int(o))), 0600)
	check(err, ERRFILE)
	return o
}

// done
func (a *Account) GenName(name string) string {
	var out []byte
	var dom string
	var dname string

	chunkOfName := strings.Split(name, ".") // slice of exploded name
	lenChunkOfName := len(chunkOfName)      // length of slice

	if lenChunkOfName > 2 { // separate domain name and domain
		dname = strings.Join(chunkOfName[:lenChunkOfName-2], "")
		dom = strings.Join(chunkOfName[lenChunkOfName-2:], "")
	} else {
		dname = strings.Join(chunkOfName[:lenChunkOfName-1], "")
		dom = strings.Join(chunkOfName[lenChunkOfName-1:], "")
	}

	//	return dom
	dnamelen := len(dname)
	domlen := len(dom)
	//	fmt.Println(domlen)
	//	fmt.Println(dnamelen)
	if dnamelen+domlen > NAMELEN {
		out = append(out, dname[:NAMELEN-domlen]...)
		out = append(out, dom...)
	} else {
		out = append(out, dname...)
		out = append(out, dom...)
	}

	return string(out)
}

// done
func (a *Account) genPass(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// done
func (a *Account) getLastUID() int {
	out, _ := exec.Command("awk",
		"-F:",
		"{uid[$3]=1}END{for(x=1100; x<=2000; x++) {if(uid[x] != \"\"){}else{print x; exit;}}}",
		"/etc/passwd",
	).Output()
	o := strings.Trim(string(out), "\n")
	ou, _ := strconv.Atoi(o)
	//	fmt.Println(out)
	//	o, _ := strconv.Atoi(string(out))
	return ou
}

// done
func (a *Account) CreateUserOs() (string, error) {
	var (
		err      error
		msg      string
		bmsg     []byte
		uid      = a.getLastUID()
		conf     = GetConfig()
		homeuser = conf.User.Home + "/" + a.Os.Name
	)

	//	fmt.Println(hom	euser, a.Os.Name)
	//	c, b := user.Lookup(a.Os.Name)
	//	fmt.Println(c, b)

	if u, er := user.Lookup(a.Os.Name); er == nil { // kalau nil maka, user sudah ada.
		//		fmt.Println(er)
		//		fmt.Printf("%+v\n", u)

		a.Os.UID, _ = strconv.Atoi(u.Uid)
		a.Os.Home = u.HomeDir
		a.Os.Pass = ""
		err = fmt.Errorf("Create user OS fail, username already taken, pick another username.")
		check(err, ERRUSR)
	} else {
		//		c := fmt.Sprintf("sudo useradd -d %s -m -s /usr/sbin/nologin -u %d -U -G docker %s", homeuser, uid, a.Os.Name)
		//		fmt.Println(c)
		//		var cmd = exec.Command("/bin/sh", "-c", c)
		var cmd = exec.Command("useradd", "-d", homeuser, "-m", "-s", "/usr/sbin/nologin", "-u", strconv.Itoa(uid), "-U", "-G", "docker", a.Os.Name)
		if _, e := cmd.Output(); e != nil {
			cmd.Stdout.Write(bmsg)
			msg = string(bmsg)
			err = e
			check(e, ERRCMD)
		} else {
			a.Os.UID = uid
			a.Os.Home = homeuser
		}
	}

	// useradd -d /sites/jualtintamcom -m -s /usr/sbin/nologin -u 1020 -U -G docker jualtintamcom
	return msg, err
}

// untested
func (a *Account) SetQuota(quota int) error {
	var softq = quota - ((quota * 4) / 100)
	// setquota -u jualtintamcom 2096640 2097152 0 0 -a
	var cmd = exec.Command("setquota", "-u", a.Os.Name, strconv.Itoa(softq), strconv.Itoa(quota), "0", "0", "-a")

	return cmd.Run()
}

// done
func (a *Account) CreateUserDb() error {
	/*
		CREATE DATABASE jualtintamcom;
		CREATE USER 'jualtintamcom'@'%' IDENTIFIED BY 'qfANXxNLCvhdh3Jm';
		GRANT ALL PRIVILEGES ON jualtintamcom.* TO 'jualtintamcom'@'%';
		FLUSH PRIVILEGES;
	*/
	//	var err error
	var exist interface{}
	var err error
	var query string
	config := GetConfig()

	db, err := sql.Open("mysql", config.Main.Dsn)
	defer db.Close()
	check(err, ERRDB)

	err = db.Ping()
	check(err, ERRDB)

	if err == nil {
		//	fmt.Printf("CREATE DATABASE IF NOT EXISTS %s\n", a.Db.Name)
		//	fmt.Printf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'\n", a.Db.User, a.Db.Pass)
		//	fmt.Printf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'\n", a.Db.Name, a.Db.User)

		query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = '%s') as 'true'", a.Db.Name)
		if debug {
			log.Println("Comitting: " + query)
		}
		rows, err := db.Query(query)

		check(err, ERRDB)
		defer rows.Close()

		rows.Next()
		err = rows.Scan(&exist)

		//	fmt.Printf("%+v\n", err)
		//	fmt.Printf("%+v\n", string(exist.([]byte)))

		if ex, _ := strconv.Atoi(string(exist.([]byte))); ex == 1 {
			a.Db.Pass = ""
		} else {
			query = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", a.Db.Name)
			if debug {
				log.Println("Commiting: " + query)
			}
			_, err = db.Exec(query)

			query = fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", a.Db.User, a.Db.Pass)
			if debug {
				log.Println("Commiting: " + query)
			}
			_, err = db.Exec(query)

			query = fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", a.Db.Name, a.Db.User)
			if debug {
				log.Println("Commiting: " + query)
			}
			_, err = db.Exec(query)

			query = "FLUSH PRIVILEGES"
			if debug {
				log.Println("Commiting: " + query)
			}
			_, err = db.Exec(query)
		}

		a.Db.Host = config.Main.Dbhost
		a.Db.Port = config.Main.Dbport
	}
	return err
}

// ssh, web, space
// done
func (a *Account) DeployAsset() error {
	var err error
	var config = GetConfig()
	//	fmt.Printf("%+v\n", config)

	for _, asset := range config.Main.Assets {
		err = os.Mkdir(path.Join(a.Os.Home, asset), 0750)
		check(err, ERRFILE)
		err = os.Chown(path.Join(a.Os.Home, asset), a.Os.UID, a.Os.UID)
		check(err, ERRFILE)
	}

	return err
}

func (a *Account) GetSshDockerfile() string {
	var out string

	/*
		FROM septianw/alpine-sshd:latest

		RUN adduser -D -s /bin/ash -u 1020 jualtintamcom
		RUN echo jualtintamcom:c7Hxenhkezted7Yn | chpasswd
	*/

	out = fmt.Sprintf("FROM septianw/alpine-sshd:latest\n\n")
	out += fmt.Sprintf("RUN adduser -D -s /bin/ash -u %d %s\n", a.Os.UID, a.Os.Name)
	out += fmt.Sprintf("RUN echo %s:%s | chpasswd", a.Os.Name, a.Ssh.Pass)

	return out
}

func (a *Account) GenPostDeployScript(postDeployCmd []string) {
	var (
		sout string
		outf string
		scrs string
		scrf string
	)

	var scr = make([]string, len(postDeployCmd)+1)
	scr[0] = "#!/bin/bash"
	for i := 1; i < len(scr); i++ {
		scr[i] = postDeployCmd[i-1]

	}
	//	log.Printf("%+v", scr)

	scrs = strings.Join(scr, "\n")
	scrf = path.Join(a.Os.Home, "postdeploy.sh")
	ioutil.WriteFile(scrf, []byte(scrs), 0700)
	err := os.Chown(scrf, a.Os.UID, a.Os.UID)
	check(err, ERRFILE)

	// write entrypoint.sh
	var nopost = []string{
		"#!/bin/bash",
		`if [ -f /run/apache2/apache2.pid ]; then rm /run/apache2/apache2.pid; fi; /usr/sbin/apache2ctl -D FOREGROUND && tail -f /var/log/apache2/error.log;`,
	}

	var post = make([]string, 3)
	for i := 0; i < 3; i++ {
		if i == 0 {
			post[i] = "#!/bin/bash"
		}

		if i == 1 {
			post[i] = "if [ -f /postdeploy.sh ]; then /postdeploy.sh && rm -f /postdeploy.sh; fi;"
		}

		if i == 2 {
			post[i] = `if [ -f /run/apache2/apache2.pid ]; then rm /run/apache2/apache2.pid; fi; /usr/sbin/apache2ctl -D FOREGROUND && tail -f /var/log/apache2/error.log;`
		}
	}

	if len(postDeployCmd) == 0 {
		sout = strings.Join(nopost, "\n")
		scrf = ""
	} else {
		sout = strings.Join(post, "\n")
	}

	outf = path.Join(a.Os.Home, "entrypoint.sh")
	ioutil.WriteFile(outf, []byte(sout), 0700)
	err = os.Chown(outf, a.Os.UID, a.Os.UID)
	check(err, ERRFILE)
}

// karena parameter tambah, di tempat lain pasti error.
// handle log dengan baik. kirim ke syslog aja. konfig syslog server dengan baik.
func (a *Account) GetWebDockerfile(platform string, platformVersion string, cmds []string, postDeployCmd []string) string {
	var out string

	a.GenPostDeployScript(postDeployCmd)

	//	switch platform {
	//	case "wordpress":
	//		out = fmt.Sprintf("FROM septianw/wordpress:%s\n\n", platformVersion)
	//	case "custom":
	//		out = fmt.Sprintf("FROM septianw/%s:%s\n\n", platform, platformVersion)
	//	default:
	//		out = fmt.Sprintf("FROM septianw/wordpress:4.5.2\n\n") // TODO: this need to be changed
	//	}
	out = fmt.Sprintf("FROM septianw/%s:%s\n\n", platform, platformVersion)

	/*
		RUN useradd -U -u 1020 -s /usr/sbin/nologin -d /var/www jualtintamcom
		RUN sed -i -e 's/export APACHE_RUN_GROUP=www-data/export APACHE_RUN_GROUP=jualtintamcom/g' /etc/apache2/envvars
		RUN sed -i -e 's/export APACHE_RUN_USER=www-data/export APACHE_RUN_USER=jualtintamcom/g' /etc/apache2/envvars
		RUN chown -Rf jualtintamcom:jualtintamcom /var/www
	*/

	//	for _, k := range adds {
	//		out += fmt.Sprintf("ADD %s %s")
	//	}
	out += fmt.Sprintf("ADD artefact /var/www/html/\n")

	out += fmt.Sprintf("RUN useradd -U -u %d -s /usr/sbin/nologin -d /var/www %s\n", a.Os.UID, a.Os.Name)
	out += fmt.Sprintf("RUN sed -i -e 's/export APACHE_RUN_GROUP=www-data/export APACHE_RUN_GROUP=%s/g' /etc/apache2/envvars\n", a.Os.Name)
	out += fmt.Sprintf("RUN sed -i -e 's/export APACHE_RUN_USER=www-data/export APACHE_RUN_USER=%s/g' /etc/apache2/envvars\n", a.Os.Name)
	out += fmt.Sprintf("RUN chown -Rf %s:%s /var/www\n", a.Os.Name, a.Os.Name)
	out += fmt.Sprintf("ADD entrypoint.sh /entrypoint.sh\n")
	if debug {
		log.Println(postDeployCmd)
	}
	if len(postDeployCmd) != 0 {
		out += fmt.Sprintf("ADD postdeploy.sh /postdeploy.sh\n")
	}
	out += fmt.Sprintf("WORKDIR /var/www/html/\n")

	// tulis deploy script
	//	fmt.Println(len(cmds))
	if len(cmds) > 0 {
		out += fmt.Sprint("\n\n")
		for _, k := range cmds {
			out += fmt.Sprintf("RUN %s\n", k)
		}
	}

	return out
}

func (a *Account) RemoveUserOs() bool {
	var balik bool
	if _, er := user.Lookup(a.Os.Name); er == nil { // kalau nil maka, user sudah ada.
		cmd := exec.Command("/usr/sbin/userdel", "-r", a.Os.Name)
		if _, e := cmd.Output(); e != nil {
			check(e, ERRCMD)
			balik = false
		} else {
			balik = true
		}
	} else {
		balik = true
	}

	return balik
}

func (a *Account) RemoveUserDb() bool {
	var (
		out   bool
		err   error
		ex    interface{}
		query string
	)
	config := GetConfig()

	db, err := sql.Open("mysql", config.Main.Dsn)
	defer db.Close()
	check(err, ERRDB)

	err = db.Ping()
	check(err, ERRDB)

	query = fmt.Sprintf("DROP USER IF EXISTS %s\n", a.Db.User) // mysql 5.7
	if debug {
		log.Println("Committing: " + query)
	}
	_, err = db.Exec(query)

	if err != nil {
		query = fmt.Sprintf("GRANT USAGE ON *.* TO '%s'@'%%' IDENTIFIED BY '%s'\n", a.Db.User, a.Db.Pass)
		if debug {
			log.Println("Committing: " + query)
		}
		_, err = db.Exec(query)
		check(err, ERRDB)

		query = fmt.Sprintf("DROP USER '%s'@'%%'\n", a.Db.User)
		if debug {
			log.Println("Committing: " + query)
		}
		_, err = db.Exec(query)
		check(err, ERRDB)
	}

	// validation

	if debug {
		log.Println("Checking if user deleted.")
	}
	q := fmt.Sprintf("select user from mysql.user where user = '%s'", a.Db.User)
	r, err := db.Query(q)
	defer r.Close()

	if err == nil {
		if r.Next() {
			e := r.Scan(&ex)
			if e != nil {
				check(e, ERRDB)
			} else {
				//				fmt.Println(string(ex.([]byte)))
				if strings.Compare(a.Db.User, string(ex.([]byte))) == 0 {
					out = false
				} else {
					out = true
				}
			}
		} else {
			out = true
		}
	} else {
		check(err, ERRDB)
		out = false
	}

	if out == true {
		if debug {
			log.Println("User deleted.")
		}
	}

	return out
}

// done
func New(domain string) *Account {
	if strings.Compare(os.Getenv("APDEBUG"), "") == 0 {
		debug = false
	} else {
		debug = true
	}

	var a = new(Account)
	var name = a.GenName(domain)
	a.Domain = domain
	a.Db.Name = name
	a.Db.User = name
	a.Os.Name = name
	a.Ssh.User = name

	a.Db.Pass = a.genPass(16)
	a.Os.Pass = a.genPass(16)
	a.Ssh.Pass = a.genPass(16)

	//	a.Os.UID = a.getLastUID()

	return a
}

package main

type DockerCompose struct {
	Version  string
	Services map[string]interface{}
}

type Web struct {
	Build         string
	Volumes       []string          `yaml:"volumes,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Restart       string
	Mem_limit     string
	Memswap_limit string
	Cpu_shares    int
	Cpu_quota     int
}

type Ssh struct {
	Build     string
	Ports     []string
	Volumes   []string `yaml:"volumes,omitempty"`
	Mem_limit string
	Restart   string `yaml:"restart,omitempty"`
}

type LoadBalancer struct {
	Image       string
	Restart     string
	Links       []string
	Volumes     []string
	Environment []string
	Ports       []string
}

/*
version: '2'
services:
  web:
    build: web
    volumes:
      - "/sites/tasmodelternet/space:/var/www/html/wp-content" # dari jalan
    environment:				# dari tool
      DBUSER: tasmodelternet	# dari tool
      DBPASS: NKgg3xwjvdxwU7qv	# dari tool
      DBHOST: 104.154.20.82		# dari tool
      DBNAME: tasmodelternet	# dari tool
    restart: always				# dari config
    mem_limit: 128M				# dari config
    memswap_limit: 1M			# dari config
    cpu_shares: 64				# dari config
    cpu_quota: 25000			# dari config
  ssh:
    build: ssh
    ports:
      - "2204:22"				# range dari config, port dari tool
    volumes:
      - "/sites/tasmodelternet/space:/home/tasmodelternet"		# dari jalan
    mem_limit: 25M				# dari config
    restart: always				# dari config
  lb:
    image: dockercloud/haproxy:1.4
    restart: always				# dari config
    links:
      - web
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - STATS_AUTH=asep:ganteng	# dari config
    ports:
      - 32788:80				# range dari config, port dari tool
      - 32789:1936				# range dari config, port dari tool
*/

type Build struct {
	Type string   `yaml:"type,omitempty"`
	Cmd  []string `yaml:"cmd,omitempty"`
}

type App struct {
	Name       string `yaml:"name,omitempty"`
	Domain     string `yaml:"domain,omitempty"`
	Platform   string
	Size       string   `yaml:"size,omitempty"`
	Type       string   `yaml:"type"`
	Storage    []string `yaml:"storage,omitempty"`
	Predeploy  []string `yaml:"predeploy,omitempty"`
	Deploy     []string `yaml:"deploy,omitempty"`
	Postdeploy []string `yaml:"postdeploy,omitempty"`
}

type Db struct {
	Engine string
	Type   string `yaml:"type,omitempty"`
	Size   string `yaml:"size,omitempty"`
}

type SuratJalan struct {
	Domain string           `yaml:"domain,omitempty"`
	Builds map[string]Build `yaml:"builds,omitempty"`
	Apps   map[string]App   `yaml:"apps"`
	Dbs    map[string]Db    `yaml:"dbs,omitempty"`
}

/*
# version: 0.1.0
# surat jalan
domain: jayatamateknik.co.id  # for custom domain, will ignore *.online.or.id

builds
  build1:  # build the apps, not node
    type: php
    cmd:
      - go build

apps:
  app1:
    name: mainapp   # used for hostname
    platform: custom
    size: small
    type: php
    storage:
      - wp-content  # basepath on current code directory
    predeploy:
      - echo predeploy
    deploy:
      - echo deploy
    postdeploy:
      - echo postdeploy

dbs:
  db1:
    engine: mysql
    type: shared
    size: small
*/

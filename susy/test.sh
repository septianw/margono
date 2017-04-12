#!/bin/bash
docker run --name jalantooldb -d -p 3307:3306 -e MYSQL_ROOT_PASSWORD=root mysql:5.7
sleep 15

go build

sudo ./jalantool go.buka.web.id
sudo userdel -r gobukawebid

docker stop jalantooldb
docker rm jalantooldb
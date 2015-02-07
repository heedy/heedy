#!/bin/bash


# create user
curl -i  --user eve:eve -H "Content-Type: application/json" -X POST -d '{"Name":"eve","Email":"eve@eve.com", "Password":"eve"}' http://localhost:8080/api/v1/json/user/
curl -i  --user eve:eve -H "Content-Type: application/json" -X POST -d '{"Name":"device1","OwnerId":"1"}' http://localhost:8080/api/v1/json/eve/device/

curl -i  --user eve:eve -H "Content-Type: application/json" -X POST -d '{"Name":"stream1","OwnerId":"1"}' http://localhost:8080/api/v1/json/eve/1/stream/

curl -i  --user eve:eve -H "Content-Type: application/json" -X POST -d '{"Timestamp":"2013-12-31T19:00:00.163523453-05:00","Data":"Hello1!"}' http://localhost:8080/api/v1/json/eve/1/1/point/
curl -i  --user eve:eve -H "Content-Type: application/json" -X POST -d '{"Timestamp":"2014-01-05T19:00:00.163523453-05:00","Data":"Hello2!"}' http://localhost:8080/api/v1/json/eve/1/1/point/
curl -i  --user eve:eve -H "Content-Type: application/json" -X POST -d '{"Timestamp":"2014-06-03T19:00:00.163523453-05:00","Data":"Hello3!"}' http://localhost:8080/api/v1/json/eve/1/1/point/

# you can now use username: eve, password: eve to access the database, or whatever API key was sent back by the device stream

echo "\n\nThe following is your device information: "

curl -i  --user eve:eve -H "Content-Type: application/json" -X GET http://localhost:8080/api/v1/xml/eve/1/ | sed -e 's/<ApiKey>\(.*\)<\/ApiKey>/\1/'

echo "\n"

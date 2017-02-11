
# GAE/Go mobile backend

* Only user management and authentication sample.

# How to create

## Create keys

```
mkdir assets
cd assets
openssl ecparam -genkey -name prime256v1 -noout -out ec256-key-pair.pem
openssl ec -in ec256-key-pair.pem -outform PEM -pubout -out ec256-key-pub.pem
openssl ec -in ec256-key-pair.pem -outform PEM -out ec256-key-pri.pem
go-bindata -o bindata/bindata.go -prefix "assets/" -pkg "bindata" assets/...
```

## Get vendor libraries

```
go get -u github.com/kardianos/govendor
govendor init
govendor fetch +out
```

# How to run development server

```
goapp serve app
```

# How to access

```
$ curl -d '{}' http://localhost:8080/user/registration

{"Success":true,"UserID":5629499534213120,"UserToken":"eee0e93e-9737-4dd6-aa43-965976d78929"}

curl -d '{"UserID":5629499534213120,"UserToken":"eee0e93e-9737-4dd6-aa43-965976d78929"}' http://localhost:8080/user/authentication

{"Success":true,"AccessToken":"eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODY3ODMwNjcsInN1YiI6NTYyOTQ5OTUzNDIxMzEyMH0.BLgQjGRitA_LYf_Qi5pwPJxAgncz9y5dYtPlmslC4A7uLqE2DrhL32Acx82MCBGF__oZ1LKRwssDXmpHQG1eeQ"}

curl -H 'Authorization: Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODY3ODMwNjcsInN1YiI6NTYyOTQ5OTUzNDIxMzEyMH0.BLgQjGRitA_LYf_Qi5pwPJxAgncz9y5dYtPlmslC4A7uLqE2DrhL32Acx82MCBGF__oZ1LKRwssDXmpHQG1eeQ' http://localhost:8080/hello

{"Success":true,"Message":"Hello 5629499534213120"}
```

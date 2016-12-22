
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

{"Success":true,"UserToken":"1818b108-cff9-487d-aa45-16e08c5e6e1f"}


$ curl -d '{"UserToken":"17c4e334-d080-4afc-adcf-cd422062ba6c"}' http://localhost:8080/user/authentication

{"Success":true,"AccessToken":"eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODIyOTI4NTMsInN1YiI6NTczMzk1MzEzODg1MTg0MH0.YTXKG7b9n7kUNJ1oXGvwOdA3-xpLDg9kkiTv3BXaqT_NoyBSKyJYp8523J6Km9OwqgRYGLR3_scI3JTYc4ojzg"}


$ curl -H 'Authorization: Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODIyOTI4NTMsInN1YiI6NTczMzk1MzEzODg1MTg0MH0.YTXKG7b9n7kUNJ1oXGvwOdA3-xpLDg9kkiTv3BXaqT_NoyBSKyJYp8523J6Km9OwqgRYGLR3_scI3JTYc4ojzg' http://localhost:8080/hello

{"Success":true,"Message":"Hello 5733953138851840"}
```

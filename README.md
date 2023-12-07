# fluffycore-starterkit-echo

## Swagger

[swag](https://github.com/swaggo/swag)  

```bash
cd cmd/server
swag init --dir .,../../internal
```

This will generate a docs.go file in the cmd/server folder.

## Downstream Services

[google](https://console.cloud.google.com/apis/credentials/oauthclient)  

## Launch Server

```powershell
cd cmd/server
go build .

$env:PORT = "9044"; .\server.exe
```

## Launch Client App

```powershell
cd cmd/clientapp2
go build .

$env:PORT = "5556";$env:GOOGLE_OAUTH2_CLIENT_ID = "1096301616546-edbl612881t7rkpljp3qa3juminskulo.apps.googleusercontent.com";$env:GOOGLE_OAUTH2_CLIENT_SECRET = "gOKwmN181CgsnQQDWqTSZjFs";$env:AUTHORITY = "http://localhost:9044"; .\clientapp2.exe
```


cd ../../cmd/stv # make sure we are in stv directory
set -x
GOOS=linux GOARCH=arm GOARM=5 go build -o stv


#node=192.168.188.120
node=120.55.42.104
fileName=dts
# 打包推送
push :
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	mv main $(fileName)
	scp $(fileName) root@$(node):/usr/local/bin
	scp -r config root@$(node):/root/
	rm -f $(fileName)

copy :
	scp -r config root@$(node):/root/

build-linux :
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	mv main $(fileName)
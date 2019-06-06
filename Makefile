default: all
all: gmon
gmon: gmon.go
	go build gmon.go
install: gmon
	install gmon /usr/local/bin

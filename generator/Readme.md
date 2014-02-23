#get steamkit submodule
git submodule update --init --recursive

#windows
build project in visual studio
go run generator.go

#linux
install monodevelop 4.x
install mono-xbuild
xbuild
go run generator.go


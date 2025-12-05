module github.com/hzcrv1911/frpcgui

go 1.24.0

require (
	github.com/Microsoft/go-winio v0.6.2
	github.com/fsnotify/fsnotify v1.9.0
	github.com/lxn/walk v0.0.0-20210112085537-c389da54e794
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e
	github.com/pelletier/go-toml/v2 v2.2.0
	github.com/samber/lo v1.47.0
	golang.org/x/sys v0.38.0
	golang.org/x/text v0.24.0
	gopkg.in/ini.v1 v1.67.0
)

require gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect

replace github.com/lxn/walk => ./libs/walk

module github.com/billylindeman/ion-client-emulator

go 1.15

// replace github.com/pion/ion-cluster => github.com/cryptagon/ion-cluster v0.0.0-20210219163655-1c8cf648e5d2
replace github.com/pion/ion-cluster => /Users/billy/Development/go/src/github.com/pion/ion-cluster

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/faiface/mainthread v0.0.0-20171120011319-8b78f0a41ae3
	github.com/giongto35/cloud-game/v2 v2.4.2-0.20210219173358-34537a5dc57d
	github.com/lucsky/cuid v1.0.2
	github.com/marten-seemann/qtls-go1-15 v0.1.1 // indirect
	github.com/pion/interceptor v0.0.9
	github.com/pion/ion-cluster v0.0.0-00010101000000-000000000000
	github.com/pion/ion-log v1.0.0
	github.com/pion/webrtc/v2 v2.2.26 // indirect
	github.com/pion/webrtc/v3 v3.0.11
	github.com/spf13/cobra v1.1.1
)

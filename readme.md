# steam-game-uninstaller

Uninstalls Steam games. Why? To uninstall games while Steam isn't running and when starting Steam without browser support (`-no-browser`), which makes Steam's uninstaller break for some reason.

Install it:
```
go get github.com/kraxarn/steam-game-uninstaller
```
If you don't want to type `steam-game-uninstaller` every time you want to uninstall something:
```
cd `go env GOPATH`/bin && ln -s ./steam-game-uninstaller ./sgu
```
Now you can run it from `sgu`
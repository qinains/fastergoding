# Fastergoding

A gopher tool for faster coding. It can automatically compile and rerun the main.main() function when the files is changed.

## Usage
```bash
cd {your-project}
go get -u github.com/qinains/fastergoding
```
then edit the `main.go` file
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/qinains/fastergoding"
)

func main() {
	fastergoding.Run() // Just add this code

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello %s!", r.URL.Query().Get("name"))
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
```

## Develop

```bash
export GOPROXY=https://goproxy.cn #if you live in China
go get -u github.com/qinains/fastergoding
```

## Reference

[https://github.com/beego/bee](https://github.com/beego/bee)

package main

import (
	"fmt"
	//	"io/ioutil"
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const debug = true

var url string = "localhost:8080"

func here(file_id string) string {
	return file_id + ".go"
}
func binpath(file_id string) string {
	return "code/" + file_id
}
func filepath(file_id string) string {
	return "code/" + file_id + ".go"
}

var beancounter int = 0

func killer(n int) {
	time.Sleep(time.Second * 2)

	if n == beancounter {
		exec.Command("killall", "go1").Run()
		exec.Command("killall", "gccgo").Run()
	}
}

func killhere(file_id string) {
	// kill the binary file
	os.OpenFile(file_id, os.O_CREATE|os.O_TRUNC, 0666)
}

func build(file_id string) bool {
	var buf bytes.Buffer

	fer := &buf

	cmd := exec.Command("go", "build", "-compiler", "gccgo", here(file_id))
	cmd.Stderr = fer
	cmd.Stdout = fer

	go killer(beancounter)

	err := cmd.Run()

	beancounter++

	if err != nil {
		fmt.Println(err)
		killhere(file_id)

		type Json struct {
			Errors string
		}
		j := Json{Errors: string(fer.Bytes())}

		xxx, err := json.Marshal(j)
		if err != nil {
			fmt.Println(err)
			return false
		}
		errorf(file_id, string(xxx), true)
		return true
	}

	return false
}

func upload(file_id string) {
	if debug {
		fmt.Println("UPLOADING\n")
	}

	postfile := file_id + ".txt"
	posturl := "http://" + url + "/u/" + file_id

	//	postfile = "../file"

	err := exec.Command("wget", `--header="Content-type: application/x-www-form-urlencoded"`, "--post-file", postfile, posturl, "-O", "-").Run()
	if err != nil {
		fmt.Println(err)
	}

}

const lerr = `{"Errors":"`
const rerr = `","Events":[{"Message":"","Kind":"stdout","Delay":0}]}`

func errorf(file_id, erro string, json bool) {
	file, err := os.Create(file_id + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	if !json {
		file.Write([]byte(lerr))
	}
	file.Write([]byte(erro))
	if !json {
		file.Write([]byte(rerr))
	}
	file.Close()
}

func compile(file_id string) {
	if debug {
		fmt.Println("compiling:", file_id)
	}
	err := os.Chdir("code/")
	if err != nil {
		fmt.Println(err)
		return
	}

	if filter(file_id) {
		if build(file_id) {
			upload(file_id)
		} else if xec(file_id) {
			upload(file_id)
		}
		killhere(file_id + ".txt")
	} else {
		os.OpenFile(file_id, os.O_CREATE|os.O_TRUNC, 0666)
		errorf(file_id, "Import is banned. Use print().", false)
		upload(file_id)
	}

	err = os.Chdir("..")
	if err != nil {
		fmt.Println(err)
	}
}
func next(ban byte) byte {
	switch ban {
	case 'i':
		return 'm'
	case 'm':
		return 'p'
	case 'p':
		return 'o'
	case 'o':
		return 'r'
	case 'r':
		return 't'
	case 't':
		return 'i'
	default:
		return 'i'
	}
}

func filter(file_id string) bool {
	f, err := os.Open(here(file_id))
	if err != nil {
		fmt.Println(err)
		return false
	}
	data := make([]byte, 1)
	var count int = 1

	ban := next(0)

	for count != 0 {
		count, err = f.Read(data)
		if err == io.EOF {
			//			fmt.Println("FOUND EOF")
			return true
		}
		if err != nil {
			fmt.Println(err)
			return false
		}
		//		fmt.Printf("%s", data[:count])

		for _, c := range data[:count] {
			//			fmt.Println("(expect %s)", ban)
			if c == ban {

				ban = next(ban)

				//				fmt.Println("(banned char %s %s)", c, ban)

				if c == 't' && ban == 'i' {
					if debug {
						fmt.Println("FOUND IMPORT")
					}
					return false
				}
			} else {
				//				fmt.Println("(normal %s %s)", c, ban)

				ban = next(0)
			}
		}
	}

	return true
}

func xec(file_id string) bool {
	file, err := os.Create(file_id + ".txt")
	if err != nil {
		fmt.Println(err)
		return false
	}
	file.Write([]byte(`{"Errors":"","Events":[{"Message":"`))

	cmd := exec.Command("./" + file_id)
	cmd.Stdout = file
	cmd.Stderr = file

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
	}
	file.Write([]byte(`","Kind":"stdout","Delay":0}]}` + "\n\n"))
	file.Close()

	killhere(file_id)
	return true
}

func download(file_id string) {
	file_name := filepath(file_id)
	res, err := http.Get("http://" + url + "/p/" + file_id + ".go")
	if err != nil {
		fmt.Println(err)
	}
	file, err := os.Create(file_name)
	if err != nil {
		fmt.Println(err)
		return
	}
	io.CopyN(file, res.Body, 131072)

	res.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	file.Close()
}

func main() {
	if len(os.Args) >= 2 {
		url = os.Args[1]
	}

	for {
		res, err := http.Get("http://" + url + "/list")
		if err != nil {
			fmt.Println(err)
		}
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			file_id := scanner.Text()
			file_name := filepath(file_id)
			bin_name := binpath(file_id)
			_, err := os.Lstat(file_name)
			if err != nil {
				download(file_id)
			}

			_, err = os.Lstat(bin_name)
			if err != nil {
				compile(file_id)
			}
			_, err = os.Lstat(bin_name)
			if err != nil {
				fmt.Println("Not compiled:", bin_name)
			}
		}
		res.Body.Close()

		time.Sleep(time.Second)
	}
}
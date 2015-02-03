package main

import (
	"fmt"
//	"io/ioutil"
	"io"
	"net/http"
	"bufio"
	"os"
	"time"
	"os/exec"
)

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

func build(file_id string) {
	err := exec.Command("go", "build", "-compiler", "gccgo", here(file_id)).Run()
	if err != nil {
		os.OpenFile(file_id, os.O_CREATE|os.O_TRUNC, 0666)
		fmt.Println(err)
	}
}

func upload(file_id string) {
	fmt.Println("UPLOADING\n")

	postfile := file_id + ".txt"
	posturl := "http://"+url+"/u/" + file_id

//	postfile = "../file"

	err := exec.Command("wget",`--header="Content-type: application/x-www-form-urlencoded"`,"--post-file",postfile,posturl,"-O","-").Run()
	if err != nil {
		fmt.Println(err)
	}

}

func errorf(file_id,erro string) {
	file, err := os.Create(file_id + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	file.Write([]byte(`{"Errors":"`))
	file.Write([]byte(erro))
	file.Write([]byte(`","Events":[{"Message":"","Kind":"stdout","Delay":0}]}`))
	file.Close()
}

func compile(file_id string) {
	fmt.Println("compiling:", file_id)

	err := os.Chdir("code/")
	if err != nil {
		fmt.Println(err)
		return
	}

	if filter(file_id) {
		build(file_id)
		if xec(file_id) {
			upload(file_id)
		}
	} else {
		os.OpenFile(file_id, os.O_CREATE|os.O_TRUNC, 0666)
		errorf(file_id, "Import is banned. Use print().")
		upload(file_id)
	}

	err = os.Chdir("..")
	if err != nil {
		fmt.Println(err)
	}
}
func next(ban byte) byte {
	switch ban {
	case 'i': return 'm';
	case 'm': return 'p';
	case 'p': return 'o';
	case 'o': return 'r';
	case 'r': return 't';
	case 't': return 'i';
	default : return 'i';
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
					fmt.Println("FOUND IMPORT")
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
	file.Write([]byte(`","Kind":"stdout","Delay":0}]}`+"\n\n"))
	file.Close()
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


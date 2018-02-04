package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/luweimy/goutil/bytefmt"

	"github.com/luweimy/gosync/workerq"
)

var (
	workerQueue = workerq.NewWorkerQueue(1)
)

func main() {
	listen := flag.String("listen", ":9293", "addr listen")
	if !flag.Parsed() {
		flag.Parse()
	}

	mux := http.NewServeMux()

	// curl -F "action=upload" -F "file=@/Users/luwei/token.md" -F "dst=/Users/luwei/tmp" -F "perm=666" http://localhost:9293/upload
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		workerQueue.AppendWorkerFuncWait(nil, func(worker *workerq.Worker) error {
			upload(w, r)
			return nil
		})
	})

	// curl -D 'ls -l' http://localhost:9293/exec
	mux.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		workerQueue.AppendWorkerFuncWait(nil, func(worker *workerq.Worker) error {
			execute(w, r)
			return nil
		})
	})
	log.Panic(http.ListenAndServe(*listen, mux))
}

func upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1024 * 1024 * 100)
	if err != nil {
		ErrorText(w, r, err)
		return
	}
	defer r.MultipartForm.RemoveAll()

	uploadPath := GetMultipartFormValue(r, "dst")
	if uploadPath == "" {
		ErrorText(w, r, "dst path not exist")
		return
	}
	perm := GetMultipartFormValue(r, "perm")
	if perm == "" {
		perm = "777"
	}
	permInt, err := strconv.ParseUint(perm, 8, 0)
	if err != nil {
		ErrorText(w, r, fmt.Errorf("perm value err(%v)", err))
		return
	}

	// 先删除目标位置的文件
	os.Remove(uploadPath)

	for _, handlers := range r.MultipartForm.File {
		for _, item := range handlers {
			fdst, err := os.OpenFile(uploadPath, os.O_CREATE|os.O_RDWR, os.FileMode(permInt))
			if err != nil {
				ErrorText(w, r, fmt.Sprintf("open dst file(%v): %s", err, uploadPath))
				return
			}
			defer fdst.Close()

			fsrc, err := item.Open()
			if err != nil {
				ErrorText(w, r, fmt.Sprintf("open src file(%v): %s", err, item.Filename))
				return
			}
			defer fsrc.Close()

			n, err := io.Copy(fdst, fsrc)
			if err != nil {
				ErrorText(w, r, fmt.Sprintf("io copy file(%v): written %d", err, n))
				return
			}

			SuccessText(w, r, fmt.Sprintf("upload  %s  %s", uploadPath, bytefmt.ByteSize(item.Size)))
		}
	}
}

func execute(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorText(w, r, err)
		return
	}

	commandArgs := strings.Fields(string(body))

	var cmd *exec.Cmd
	if len(commandArgs) < 2 {
		cmd = exec.Command(commandArgs[0])
	} else {
		cmd = exec.Command(commandArgs[0], commandArgs[1:]...)
	}
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		ErrorText(w, r, err)
		return
	}

	SuccessText(w, r, string(output))
}

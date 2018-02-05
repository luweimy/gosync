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
	"github.com/luweimy/goutil/workerq"
)

var (
	workerQueue *workerq.WorkerQueue
)

func main() {
	listen := flag.String("listen", ":9293", "address listen")
	concurrency := flag.Int("concurrency", 1, "concurrency limit")
	if !flag.Parsed() {
		flag.Parse()
	}

	// 创建WorkerQueue
	workerQueue = workerq.NewWorkerQueue(*concurrency)

	mux := http.NewServeMux()

	// curl -F "file=@/Users/luwei/token.md" -F "dst=/Users/luwei/tmp" -F "perm=666" http://localhost:9293/upload
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		worker := workerQueue.AppendWorkerFunc(nil, func(worker *workerq.Worker) error {
			upload(w, r)
			return nil
		})
		if err := worker.Wait(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(fmt.Sprintf("\n%s %s %s done.\n", NowString(), r.Method, r.URL)))
	})

	// curl -d 'ls -l' http://localhost:9293/exec
	mux.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		worker := workerQueue.AppendWorkerFunc(nil, func(worker *workerq.Worker) error {
			execute(w, r)
			return nil
		})
		if err := worker.Wait(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(fmt.Sprintf("\n%s %s %s done.\n", NowString(), r.Method, r.URL)))
	})

	// curl -F "n=10" http://localhost:9293/concurrency
	mux.HandleFunc("/concurrency", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(0); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		n, err := strconv.ParseInt(r.PostForm.Get("n"), 10, 0)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		workerQueue.SetConcurrency(int(n))
		w.Write([]byte(fmt.Sprintf("\n%s %s %s done.\n", NowString(), r.Method, r.URL)))
	})

	log.Panic(http.ListenAndServe(*listen, mux))
}

func upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1024 * 1024 * 100)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer r.MultipartForm.RemoveAll()

	uploadPath := GetMultipartFormValue(r, "dst")
	if uploadPath == "" {
		w.Write([]byte("dst path not exist"))
		return
	}
	perm := GetMultipartFormValue(r, "perm")
	if perm == "" {
		perm = "777"
	}
	permInt, err := strconv.ParseUint(perm, 8, 0)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	// 先删除目标位置的文件
	os.Remove(uploadPath)

	for _, handlers := range r.MultipartForm.File {
		for _, item := range handlers {
			fdst, err := os.OpenFile(uploadPath, os.O_CREATE|os.O_RDWR, os.FileMode(permInt))
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			defer fdst.Close()

			fsrc, err := item.Open()
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			defer fsrc.Close()

			_, err = io.Copy(fdst, fsrc)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}

			w.Write([]byte(fmt.Sprintf("upload  %s  %s", uploadPath, bytefmt.ByteSize(item.Size))))
		}
	}
}

func execute(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
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
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(string(output)))
}

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"
	"time"
)

var requestCounter uint64

type dumpJob struct {
	filename string
	data     []byte
}

var dumpChan chan dumpJob

func dumpWorker() {
	for job := range dumpChan {
		os.WriteFile(job.filename, job.data, 0644)
	}
}

func main() {
	targetUrl := flag.String("origin", "", "Target URL")
	port := flag.Int("p", 8000, "Port")
	dump := flag.Bool("dump", false, "Dump each request/response to file in current dir")
	flag.StringVar(targetUrl, "o", "https://api.ppinfra.com", "Target URL (shorthand)")
	flag.Parse()
	if *dump {
		dumpChan = make(chan dumpJob, 100)
		go dumpWorker()
	}

	target, _ := url.Parse(*targetUrl)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqID := atomic.AddUint64(&requestCounter, 1)
		timestamp := time.Now().Format("20060102_150405")

		// ---- 打印请求 ----
		reqDump, _ := httputil.DumpRequest(r, true)
		fmt.Println("\n=== Incoming Request ===")
		fmt.Println(string(reqDump))

		// ---- 构建新的转发请求 ----
		// 用原始 method 和 body
		outReq, err := http.NewRequest(r.Method, target.String()+r.URL.Path, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// 拷贝 headers
		outReq.Header = r.Header.Clone()

		// ---- 发往上游 ----
		resp, err := http.DefaultClient.Do(outReq)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer resp.Body.Close()

		// ---- 打印响应 ----
		fmt.Println("=== Upstream Response ===")
		respDump, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(respDump))

		if *dump {
			data := append(reqDump, []byte("\n\n---\n\n")...)
			data = append(data, respDump...)
			dumpChan <- dumpJob{fmt.Sprintf("%s_%d_%d.txt", timestamp, reqID, resp.StatusCode), data}
		}

		// ---- 回传响应 ----
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	fmt.Println("Forward proxy running on", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), handler))
}

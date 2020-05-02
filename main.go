package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func handleError(w http.ResponseWriter, format string, args ...interface{}) {
	w.WriteHeader(500)
	log.Fatalf(format, args...)
	fmt.Fprintf(w, format, args...)
}

func handler(repoRoot string, resp http.ResponseWriter, req *http.Request) {
	treeish := req.URL.Path[len("/archive/"):]
	treeish = strings.TrimSuffix(treeish, ".tar.zstd")

	gitCmd := exec.Command("git", "-C", repoRoot, "archive", "--", treeish)
	gitStdout, err := gitCmd.StdoutPipe()
	defer func() { gitStdout.Close(); gitCmd.Wait() }()

	if err != nil {
		handleError(resp, "Cannot open pipe to git-archive: %v", err)
		return
	}
	if err := gitCmd.Start(); err != nil {
		handleError(resp, "Cannot execute git-archive: %v", err)
		return
	}
	zstdCmd := exec.Command("zstd", "-")
	zstdCmd.Stdin = gitStdout
	zstdStdout, err := zstdCmd.StdoutPipe()
	defer func() { zstdStdout.Close(); zstdCmd.Wait() }()

	if err != nil {
		handleError(resp, "Cannot open pipe from zstd: %v", err)
		return
	}
	if err := zstdCmd.Start(); err != nil {
		handleError(resp, "Cannot execute zstdCmd: %v", err)
		return
	}

	resp.Header().Set("Content-Type", "application/zstd")
	if _, err := io.Copy(resp, zstdStdout); err != nil {
		handleError(resp, "Pipe failed: %v", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s git-repo\n", os.Args[0])
		os.Exit(1)
	}
	http.HandleFunc("/archive/", func(w http.ResponseWriter, r *http.Request) {
		handler(os.Args[1], w, r)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

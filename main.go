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
	log.Printf(format, args...)
	fmt.Fprintf(w, format, args...)
}

func handler(repoRoot string, resp http.ResponseWriter, req *http.Request) {
	treeish := req.URL.Path[len("/archive/"):]
	treeish = strings.TrimSuffix(treeish, ".tar.zstd")

	cmd := exec.Command("git",
		"-C", repoRoot,
		"-c", "tar.tar.zstd.command=zstd",
		"archive",
		"--format", "tar.zstd",
		"--", treeish)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		handleError(resp, "Cannot open pipe to git-archive: %v", err)
		return
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		handleError(resp, "Cannot execute git-archive: %v", err)
		return
	}

	resp.Header().Set("Content-Type", "application/zstd")
	if _, err := io.Copy(resp, stdout); err != nil {
		handleError(resp, "Pipe failed: %v", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		handleError(resp, "git-archive failed: %v", err)
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

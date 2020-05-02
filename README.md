# git-serve-zstd

Provide a repository export server for serving [zstd](https://facebook.github.io/zstd/) tarballs via http

## USAGE

```bash
go build .
./git-archive-zstd ../path/to/repo
```

```bash
wget http://localhost:8080/archive/master.tar.zstd
```

with import <nixpkgs> {};
stdenv.mkDerivation {
  name = "env";
  buildInputs = [
    bashInteractive
    go
    zstd
    git
  ];
  shellHook = ''
    unset GOROOT GOPATH
  '';
}

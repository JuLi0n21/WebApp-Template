{
  description = "Dev environment for Horsa";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }:
    let
      pkgs = import nixpkgs { system = "x86_64-linux"; };
    in
    {
      devShells.x86_64-linux.default = pkgs.mkShell {
        packages = with pkgs; [
          go_1_25
          sqlc
          protobuf
          protoc-gen-go
          protoc-gen-go-grpc
          openapi-generator-cli

          nodejs_20
          nodePackages.npm

          git
          gnumake
        ];

        env = {
          PROTOC = "${pkgs.protobuf}/bin/protoc";
          PROTOC_GEN_GO = "${pkgs.protoc-gen-go}/bin/protoc-gen-go";
          PROTOC_GEN_GO_GRPC = "${pkgs.protoc-gen-go-grpc}/bin/protoc-gen-go-grpc";
          OPENAPI_GENERATOR = "${pkgs.openapi-generator-cli}/bin/openapi-generator-cli";

          GOFLAGS = "-buildvcs=false";
        };

        shellHook = ''
          go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
          go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
        '';
      };
    };
}

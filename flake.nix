{
  description = "A simple NixOS cache server using Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: 
    flake-utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ] (system:
    let
      pkgs = import nixpkgs { inherit system; };
    in
    {
      packages.default = pkgs.stdenv.mkDerivation {
        pname = "go-cache-server";
        version = "0.1.0";
        src = ./.;
        buildInputs = [ pkgs.go ];

        # Set GOCACHE environment variable to a writable directory
        buildPhase = ''
          export HOME=$(pwd)
          export GOCACHE=$(mktemp -d)
          export GOPROXY=direct

          mkdir -p $GOCACHE
          cd src
          go build -o $out/bin/cache-server main.go
        '';

        installPhase = ''
          mkdir -p $out/bin
          cp ./config.json $out/bin/
        '';

        meta = with pkgs.lib; {
          description = "A simple NixOS cache server using Go";
          license = licenses.mit;
          maintainers = [ ];
        };
      };

      defaultPackage = self.packages.${system}.default;

      devShell = pkgs.mkShell {
        buildInputs = [ pkgs.go ];
      };

      # NixOS module for running the server as a service with Nginx and Certbot
      nixosModules.cacheServer = {
        config, pkgs, ... }: {
          services.cache-server = {
            enable = true;
            package = self.packages.${system}.default;
          };

          systemd.services.cache-server = {
            description = "Go Cache Server";
            after = [ "network.target" ];
            wantedBy = [ "multi-user.target" ];

            serviceConfig = {
              ExecStart = "${self.packages.${system}.default}/bin/cache-server";
              Restart = "always";
              WorkingDirectory = "/var/cache/go-cache-server";
            };
          };

          networking.firewall.allowedTCPPorts = [ 80 443 ];

          # Ensure the cache directory exists
          systemd.tmpfiles.rules = [
            "d /var/cache/go-cache-server 0755 root root -"
          ];

          services.nginx = {
            enable = true;
            virtualHosts."rasp.local" = {
              enableACME = true;
              forceSSL = true;
              useACMEHost = "rasp.local";
              locations."/" = {
                proxyPass = "http://127.0.0.1:8080";
              };
            };
          };

          security.acme.certs."rasp.local" = {
            email = "your-email@example.com"; # Replace with your email
            webroot = "/var/www";
          };
        };
    });
}

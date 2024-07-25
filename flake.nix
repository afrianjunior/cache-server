{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation {
  name = "cache-server";
  builder = ./builder.sh;
  args = [ ./config.json ];
  buildInputs = with pkgs; [ python nginx letsencrypt ];
}


# nginx.conf


# letsencrypt.conf
letsencrypt = {
  domains = [ "rasp.local" ];
  email = "your_email@example.com";
  agreeTos = true;
};

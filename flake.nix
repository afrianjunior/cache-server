{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation {
  name = "cache-server";
  builder = ./builder.sh;
  args = [ ./config.json ];
  buildInputs = with pkgs; [ go ];
}

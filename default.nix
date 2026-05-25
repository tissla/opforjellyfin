{ lib, stdenv, buildGoModule, fetchFromGitHub, git }:

buildGoModule rec {
  pname = "opforjellyfin";
  version = "1.0.1";

  src = fetchFromGitHub {
    owner = "tissla";
    repo = "opforjellyfin";
    rev = "d6133d89836c9c727438794ab634f6a2b1184540";
    hash = "sha256-lSB+F7heenXEmr6T+PKTRC9ZLEPGMcG5nEtVTzJUe+A=";
  };

  vendorHash = "sha256-PL42t4SywbXmpPtetau03AsTHAGmhOrajsSyF4LJwUU=";
  nativeBuildInputs = [ git ];
  propagatedBuildInputs = [ git ];

  postInstall = ''
    mv $out/bin/opforjellyfin $out/bin/opfor
  '';

  meta = with lib; {
    description = "CLI to automate download and organisation of One Pace episodes for Jellyfin";
    homepage = "https://github.com/tissla/opforjellyfin";
    license = licenses.gpl3Plus;
    platforms = platforms.linux;
    maintainers = [ nahieluniversal ];
  };
}

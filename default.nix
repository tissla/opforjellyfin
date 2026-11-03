{ lib, stdenv, buildGoModule, fetchFromGitHub, git }:

buildGoModule rec {
  pname = "opforjellyfin";
  version = "1.1.0";

  src = fetchFromGitHub {
    owner = "tissla";
    repo = "opforjellyfin";
    rev = "7bb22b7b547256366ee83ddfb6e9360eb02603a2";
    hash = "sha256-FlSc776Q6DLLBWGYkxCg1vkJlyLm3dkEtGsLbn3b20w=";
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

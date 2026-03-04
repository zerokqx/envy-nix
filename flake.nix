{
  description = "Envy flake";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
  let
    systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
  in
  {
    packages = forAllSystems (system:
      let
        pkgs = import nixpkgs { inherit system; };
        envy = pkgs.buildGoModule {
          pname = "envy";
          # если есть теги — можно руками держать 1.2.1, либо так:
          version = self.shortRev or "dirty";

          # берём исходники из текущего репозитория
          src = self;

          # у тебя уже найденный vendorHash — можно оставить.
          # если поменяешь go.mod/go.sum — обновляй (nix build подскажет got:)
          vendorHash = "sha256-1w9NbifOpZ+yvQPfOVyj0khXlN7MD+j1dznqZRVv66c=";

          # у тебя main судя по всему в ./cmd
          subPackages = [ "cmd" ];

          # если хочешь чтобы бинарник был envy, а не cmd — включи это:
          # postInstall = ''
          #   if [ -f "$out/bin/cmd" ] && [ ! -f "$out/bin/envy" ]; then
          #     mv "$out/bin/cmd" "$out/bin/envy"
          #   fi
          # '';
        };
      in {
        default = envy;
        envy = envy;
      }
    );

    apps = forAllSystems (system: {
      default = {
        type = "app";
        # если оставляем имя cmd:
        program = "${self.packages.${system}.default}/bin/cmd";
        # если включишь postInstall с переименованием, поменяй на /bin/envy
      };
    });
  };
}

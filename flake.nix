{
  description = "Envy flake";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };

          envy = pkgs.buildGoModule {
            pname = "envy";
            version = self.shortRev or "dirty";

            src = self;

            vendorHash = "sha256-1w9NbifOpZ+yvQPfOVyj0khXlN7MD+j1dznqZRVv66c=";

            subPackages = [ "cmd" ];

            postInstall = ''
              if [ -f "$out/bin/cmd" ] && [ ! -f "$out/bin/envy" ]; then
                mv "$out/bin/cmd" "$out/bin/envy"
              fi
            '';
          };
        in
        {
          default = envy;
          envy = envy;
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/envy";
        };
      });
    };
}

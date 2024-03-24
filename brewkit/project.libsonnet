local images = import 'images.libsonnet';
local schemas = import 'schemas.libsonnet';

local cache = std.native('cache');
local copy = std.native('copy');
local copyFrom = std.native('copyFrom');

// External cache for go compiler, go mod, golangci-lint
local gocache = [
    cache("go-build", "/app/cache"),
    cache("go-mod", "/go/pkg/mod"),
];

// Sources which will be tracked for changes
local gosources = [
    "go.mod",
    "go.sum",
    "cmd",
    "internal",
];

{
    project(appIDs):: {
        apiVersion: "brewkit/v2",

        targets: {
            all: ['build', 'test', 'check'],

            // build target to chain all build of apps
            build: [appID for appID in appIDs],
        } + {
            [appID]: {
                from: "_gobase",
                workdir: "/app",
                cache: gocache,
                copy: [
                    // copy go.mod changes
                    copyFrom(
                        'gotidy',
                        '/app/go.*',
                        '.'
                    )
                ],
                command: 'go build \\
                     -trimpath -v \\
                     -ldflags "-X main.AppVersion=${revision}" \\
                     -o ./bin/' + appID + ' ./cmd/' + appID,
                output: std.format("/app/bin/%s:./bin/", [appID]),
            }
            for appID in appIDs // expand build target for each appID
        } + {
            modules: ['gotidy', 'modulesvendor'],

            gotidy: {
                from: "_gobase",
                workdir: "/app",
                cache: gocache,
                ssh: {},
                command: "go mod tidy",
                output: {
                    artifact: "/app/go.*",
                    "local": ".",
                },
            },

            // export local copy of dependencies for ide index
            modulesvendor: {
                from: "_gobase",
                workdir: "/app",
                cache: gocache,
                command: "go mod vendor",
                output: {
                    artifact: "/app/vendor",
                    "local": "vendor",
                },
            },

            test: {
                from: "_gobase",
                workdir: "/app",
                cache: gocache,
                command: "go test ./...",
            },

            check: {
                from: images.golangcilint,
                workdir: "/app",
                env: {
                    GOCACHE: "/app/cache/go-build",
                    GOLANGCI_LINT_CACHE: "/app/cache/go-build",
                },
                cache: gocache,
                copy: [
                    copy('.golangci.yml', '.golangci.yml'),
                    copyFrom(
                        "_gosources",
                        "/app",
                        "/app",
                    )
                ],
                command: "golangci-lint run",
            },

            _gosources: {
                from: "scratch",
                workdir: "/app",
                copy: [copy(source, source) for source in gosources]
            },

            _gobase: {
                from: images.go,
                workdir: "/app",
                env: {
                    GOCACHE: "/app/cache/go-build",
                    CGO_ENABLED: "0",
                },
                copy: copyFrom(
                    "_gosources",
                    "/app",
                    "/app",
                ),
            },
        },
    },
}

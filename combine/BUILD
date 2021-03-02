load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")


go_binary(
    name = "main_combine",
    srcs = ["main_combine.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_apache_beam//sdks/go/pkg/beam:go_default_library",
        "@com_github_apache_beam//sdks/go/pkg/beam/log:go_default_library",
        "@com_github_apache_beam//sdks/go/pkg/beam/io/textio:go_default_library",
        "@com_github_apache_beam//sdks/go/pkg/beam/x/beamx:go_default_library",
    ],
)
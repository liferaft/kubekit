package manifest

// Version is the current KubeKit and KubeOS version
const Version = "2.1.0"

// Prerelease is a marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var Prerelease = ""

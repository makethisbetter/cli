const PLATFORM_PACKAGES = {
  darwin: {
    arm64: "@makethisbetter/cli-darwin-arm64",
    x64: "@makethisbetter/cli-darwin-x64",
  },
  linux: {
    arm64: "@makethisbetter/cli-linux-arm64",
    x64: "@makethisbetter/cli-linux-x64",
  },
  win32: {
    x64: "@makethisbetter/cli-win32-x64",
  },
};

const RELEASE_PLATFORMS = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};
const RELEASE_ARCHITECTURES = { arm64: "arm64", x64: "amd64" };

function binaryExtension(platform) {
  return platform === "win32" ? ".exe" : "";
}

function packageName(platform, architecture) {
  const packages = PLATFORM_PACKAGES[platform];
  return packages && packages[architecture];
}

function packageBinarySpecifier(platform, architecture) {
  const name = packageName(platform, architecture);
  if (!name) return undefined;

  return `${name}/bin/makethisbetter${binaryExtension(platform)}`;
}

function releaseBinaryName(platform, architecture) {
  const releasePlatform = RELEASE_PLATFORMS[platform];
  const releaseArchitecture = RELEASE_ARCHITECTURES[architecture];
  if (!releasePlatform || !releaseArchitecture) return undefined;

  return `makethisbetter-${releasePlatform}-${releaseArchitecture}${binaryExtension(platform)}`;
}

module.exports = {
  PLATFORM_PACKAGES,
  binaryExtension,
  packageBinarySpecifier,
  packageName,
  releaseBinaryName,
};

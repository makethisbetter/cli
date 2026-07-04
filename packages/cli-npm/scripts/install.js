const fs = require("fs");
const path = require("path");

const VERSION = require("../package.json").version;

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

const PLATFORM_MAP = { darwin: "darwin", linux: "linux", win32: "windows" };
const ARCH_MAP = { arm64: "arm64", x64: "amd64" };

// Check if the binary is already available via optionalDependencies
try {
  var pkgs = PLATFORM_PACKAGES[process.platform];
  if (pkgs) {
    require.resolve(pkgs[process.arch] + "/bin/makethisbetter");
    process.exit(0);
  }
} catch {}

var platform = PLATFORM_MAP[process.platform];
var arch = ARCH_MAP[process.arch];

if (!platform || !arch) {
  console.error(
    "@makethisbetter/cli postinstall: unsupported platform " +
      process.platform +
      "/" +
      process.arch +
      ", skipping binary download"
  );
  process.exit(0);
}

var ext = process.platform === "win32" ? ".exe" : "";
var filename = "makethisbetter-" + platform + "-" + arch + ext;
var url =
  "https://github.com/makethisbetter/cli/releases/download/v" +
  VERSION +
  "/" +
  filename;

var destDir = path.join(__dirname, "..", "downloaded");
fs.mkdirSync(destDir, { recursive: true });
var dest = path.join(destDir, "makethisbetter" + ext);

function download(url, dest, redirects) {
  if (redirects > 5) {
    console.error("@makethisbetter/cli postinstall: too many redirects");
    process.exit(0);
  }

  var mod = url.startsWith("https") ? require("https") : require("http");

  mod
    .get(url, { headers: { "User-Agent": "makethisbetter-cli" } }, function (res) {
      if (
        res.statusCode >= 300 &&
        res.statusCode < 400 &&
        res.headers.location
      ) {
        return download(res.headers.location, dest, redirects + 1);
      }

      if (res.statusCode !== 200) {
        console.error(
          "@makethisbetter/cli postinstall: download failed with status " +
            res.statusCode
        );
        process.exit(0);
      }

      var file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on("finish", function () {
        file.close();
        fs.chmodSync(dest, 0o755);
      });
    })
    .on("error", function () {
      // Silent fail — user can install via go install
      process.exit(0);
    });
}

download(url, dest, 0);

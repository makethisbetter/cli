const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const test = require("node:test");
const {
  packageBinarySpecifier,
  packageName,
  releaseBinaryName,
} = require("../scripts/platform");

const PACKAGE_ROOT = path.resolve(__dirname, "..");
const PLATFORM_PACKAGE_ROOTS = [
  "cli-npm-darwin-arm64",
  "cli-npm-darwin-x64",
  "cli-npm-linux-arm64",
  "cli-npm-linux-x64",
  "cli-npm-win32-x64",
].map((name) => path.resolve(PACKAGE_ROOT, "..", name));

const PLATFORM_CONTRACTS = [
  {
    directory: "cli-npm-darwin-arm64",
    platform: "darwin",
    architecture: "arm64",
    packageName: "@makethisbetter/cli-darwin-arm64",
    binary: "bin/makethisbetter",
    releaseBinary: "makethisbetter-darwin-arm64",
  },
  {
    directory: "cli-npm-darwin-x64",
    platform: "darwin",
    architecture: "x64",
    packageName: "@makethisbetter/cli-darwin-x64",
    binary: "bin/makethisbetter",
    releaseBinary: "makethisbetter-darwin-amd64",
  },
  {
    directory: "cli-npm-linux-arm64",
    platform: "linux",
    architecture: "arm64",
    packageName: "@makethisbetter/cli-linux-arm64",
    binary: "bin/makethisbetter",
    releaseBinary: "makethisbetter-linux-arm64",
  },
  {
    directory: "cli-npm-linux-x64",
    platform: "linux",
    architecture: "x64",
    packageName: "@makethisbetter/cli-linux-x64",
    binary: "bin/makethisbetter",
    releaseBinary: "makethisbetter-linux-amd64",
  },
  {
    directory: "cli-npm-win32-x64",
    platform: "win32",
    architecture: "x64",
    packageName: "@makethisbetter/cli-win32-x64",
    binary: "bin/makethisbetter.exe",
    releaseBinary: "makethisbetter-windows-amd64.exe",
  },
];

function runNpm(npmCli, args, options = {}) {
  const result = spawnSync(process.execPath, [npmCli, ...args], {
    encoding: "utf8",
    ...options,
  });

  assert.equal(
    result.status,
    0,
    [result.stdout, result.stderr].filter(Boolean).join("\n")
  );
  return result;
}

function pack(npmCli, packageDir, destination, env) {
  const result = runNpm(
    npmCli,
    ["pack", packageDir, "--pack-destination", destination, "--json"],
    { env }
  );
  return path.join(destination, JSON.parse(result.stdout)[0].filename);
}

function writePlatformFixture(source, destination) {
  fs.cpSync(source, destination, { recursive: true });
  const manifest = JSON.parse(
    fs.readFileSync(path.join(destination, "package.json"), "utf8")
  );
  const extension = manifest.os.includes("win32") ? ".exe" : "";
  const binaryPath = path.join(
    destination,
    "bin",
    `makethisbetter${extension}`
  );
  fs.mkdirSync(path.dirname(binaryPath), { recursive: true });
  fs.writeFileSync(
    binaryPath,
    "#!/usr/bin/env node\nconsole.log('fixture help');\n"
  );
  fs.chmodSync(binaryPath, 0o755);
}

function buildPackageFixture(npmCli, prefix, envOverrides = {}) {
  const sandbox = fs.mkdtempSync(path.join(os.tmpdir(), prefix));
  const packagesDir = path.join(sandbox, "packages");
  const tarballsDir = path.join(sandbox, "tarballs");
  const cacheDir = path.join(sandbox, "cache");
  const binDir = path.join(sandbox, "bin");
  fs.mkdirSync(packagesDir);
  fs.mkdirSync(tarballsDir);
  fs.mkdirSync(binDir);
  fs.symlinkSync(process.execPath, path.join(binDir, "node"));

  const env = {
    ...process.env,
    PATH: [binDir, "/usr/bin", "/bin", "/usr/sbin", "/sbin"].join(
      path.delimiter
    ),
    npm_config_cache: cacheDir,
    npm_config_update_notifier: "false",
    ...envOverrides,
  };

  try {
    const optionalDependencies = {};
    for (const source of PLATFORM_PACKAGE_ROOTS) {
      const destination = path.join(packagesDir, path.basename(source));
      writePlatformFixture(source, destination);
      const manifest = JSON.parse(
        fs.readFileSync(path.join(destination, "package.json"), "utf8")
      );
      optionalDependencies[manifest.name] = `file:${pack(
        npmCli,
        destination,
        tarballsDir,
        env
      )}`;
    }

    const wrapperDir = path.join(packagesDir, "cli-npm");
    fs.cpSync(PACKAGE_ROOT, wrapperDir, { recursive: true });
    const wrapperManifestPath = path.join(wrapperDir, "package.json");
    const wrapperManifest = JSON.parse(
      fs.readFileSync(wrapperManifestPath, "utf8")
    );
    wrapperManifest.optionalDependencies = optionalDependencies;
    fs.writeFileSync(
      wrapperManifestPath,
      `${JSON.stringify(wrapperManifest, null, 2)}\n`
    );

    return {
      sandbox,
      cacheDir,
      env,
      wrapperTarball: pack(npmCli, wrapperDir, tarballsDir, env),
    };
  } catch (error) {
    fs.rmSync(sandbox, { recursive: true, force: true });
    throw error;
  }
}

test("platform manifests match wrapper and release binary contracts", () => {
  const wrapperManifest = require("../package.json");
  const expectedPackages = PLATFORM_CONTRACTS.map(({ packageName }) =>
    packageName
  ).sort();
  assert.deepEqual(
    Object.keys(wrapperManifest.optionalDependencies).sort(),
    expectedPackages
  );

  for (const contract of PLATFORM_CONTRACTS) {
    const manifest = require(`../../${contract.directory}/package.json`);
    assert.equal(manifest.name, contract.packageName);
    assert.deepEqual(manifest.os, [contract.platform]);
    assert.deepEqual(manifest.cpu, [contract.architecture]);
    assert.equal(
      packageName(contract.platform, contract.architecture),
      contract.packageName
    );
    assert.equal(
      packageBinarySpecifier(contract.platform, contract.architecture),
      `${contract.packageName}/${contract.binary}`
    );
    assert.equal(
      releaseBinaryName(contract.platform, contract.architecture),
      contract.releaseBinary
    );
  }
});

test("published package runs through npm exec and npx without a global CLI", {
  skip: process.platform === "win32",
}, () => {
  const npmCli = process.env.NPM_CLI_JS || process.env.npm_execpath;
  assert.ok(npmCli, "set NPM_CLI_JS to the npm-cli.js under test");

  const { sandbox, cacheDir, env, wrapperTarball } = buildPackageFixture(
    npmCli,
    "mtb-cli-install-"
  );

  try {
    const npmExec = runNpm(
      npmCli,
      [
        "exec",
        "--yes",
        `--package=file:${wrapperTarball}`,
        "--",
        "makethisbetter",
        "--help",
      ],
      { cwd: sandbox, env }
    );
    assert.match(npmExec.stdout, /fixture help/);

    const npxCli = path.join(path.dirname(npmCli), "npx-cli.js");
    const npx = runNpm(
      npxCli,
      ["-y", `file:${wrapperTarball}`, "--help"],
      {
        cwd: sandbox,
        env: { ...env, npm_config_cache: `${cacheDir}-npx` },
      }
    );
    assert.match(npx.stdout, /fixture help/);

    const globalPrefix = path.join(sandbox, "global");
    runNpm(
      npmCli,
      [
        "install",
        "--global",
        "--prefix",
        globalPrefix,
        `file:${wrapperTarball}`,
      ],
      {
        cwd: sandbox,
        env: { ...env, npm_config_cache: `${cacheDir}-global` },
      }
    );
    const globalRun = spawnSync(
      path.join(globalPrefix, "bin", "makethisbetter"),
      ["--help"],
      { encoding: "utf8", env }
    );
    assert.equal(
      globalRun.status,
      0,
      [globalRun.stdout, globalRun.stderr].filter(Boolean).join("\n")
    );
    assert.match(globalRun.stdout, /fixture help/);
  } finally {
    fs.rmSync(sandbox, { recursive: true, force: true });
  }
});

test("Windows wrapper uses its installed optional binary without postinstall", {
  skip: process.platform === "win32",
}, () => {
  const npmCli = process.env.NPM_CLI_JS || process.env.npm_execpath;
  assert.ok(npmCli, "set NPM_CLI_JS to the npm-cli.js under test");

  const { sandbox, env, wrapperTarball } = buildPackageFixture(
    npmCli,
    "mtb-cli-windows-",
    {
      npm_config_cpu: "x64",
      npm_config_ignore_scripts: "true",
      npm_config_os: "win32",
    }
  );
  const installDir = path.join(sandbox, "install");

  try {
    runNpm(
      npmCli,
      ["install", "--prefix", installDir, `file:${wrapperTarball}`],
      { cwd: sandbox, env }
    );

    const preload = path.join(sandbox, "windows-platform.js");
    fs.writeFileSync(
      preload,
      [
        'Object.defineProperty(process, "platform", { value: "win32" });',
        'Object.defineProperty(process, "arch", { value: "x64" });',
        "",
      ].join("\n")
    );
    const wrapper = path.join(
      installDir,
      "node_modules",
      "@makethisbetter",
      "cli",
      "bin",
      "makethisbetter"
    );
    const result = spawnSync(
      process.execPath,
      ["--require", preload, wrapper, "--help"],
      { encoding: "utf8", env }
    );
    assert.equal(
      result.status,
      0,
      [result.stdout, result.stderr].filter(Boolean).join("\n")
    );
    assert.match(result.stdout, /fixture help/);
  } finally {
    fs.rmSync(sandbox, { recursive: true, force: true });
  }
});

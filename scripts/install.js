const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { execFileSync } = require("child_process");
const os = require("os");

const VERSION = require("../package.json").version;
const REPO = "tcl-eaglelab/tcl-smart-home-cli";
const NAME = "tcl";

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const platform = PLATFORM_MAP[process.platform];
let arch = ARCH_MAP[process.arch];

if (platform === "windows" && process.arch === "arm64") {
  console.log("Windows ARM64 detected, falling back to x64 binary (runs via emulation)");
  arch = "amd64";
}

if (!platform || !arch) {
  console.error(
    `Unsupported platform: ${process.platform}-${process.arch}`
  );
  process.exit(1);
}

const isWindows = process.platform === "win32";
const ext = isWindows ? ".zip" : ".tar.gz";
const archiveName = `${NAME}-${VERSION}-${platform}-${arch}${ext}`;
const GITHUB_URL = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;

const binDir = path.join(__dirname, "..", "bin");
const dest = path.join(binDir, NAME + (isWindows ? ".exe" : ""));

fs.mkdirSync(binDir, { recursive: true });

function download(url, destPath) {
  const args = [
    "--fail", "--location", "--silent", "--show-error",
    "--connect-timeout", "10", "--max-time", "120",
    "--output", destPath, url,
  ];
  if (isWindows) {
    args.unshift("--ssl-revoke-best-effort");
  }
  execFileSync("curl", args, { stdio: ["ignore", "ignore", "pipe"] });
}

function verifyChecksum(filePath, expectedHash) {
  const fileBuffer = fs.readFileSync(filePath);
  const hash = crypto.createHash("sha256").update(fileBuffer).digest("hex");
  if (hash !== expectedHash) {
    throw new Error(
      `Checksum mismatch!\n  Expected: ${expectedHash}\n  Actual:   ${hash}`
    );
  }
}

function install() {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "tcl-cli-"));
  const archivePath = path.join(tmpDir, archiveName);
  const checksumURL = `https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt`;
  const checksumPath = path.join(tmpDir, "checksums.txt");

  try {
    download(GITHUB_URL, archivePath);

    try {
      download(checksumURL, checksumPath);
      const checksumContent = fs.readFileSync(checksumPath, "utf8");
      const line = checksumContent
        .split("\n")
        .find((l) => l.includes(archiveName));
      if (line) {
        const expectedHash = line.trim().split(/\s+/)[0];
        verifyChecksum(archivePath, expectedHash);
        console.log("Checksum verified");
      } else {
        console.warn("Warning: archive not found in checksums.txt, skipping verification");
      }
    } catch (checksumErr) {
      if (checksumErr.message.includes("Checksum mismatch")) {
        throw checksumErr;
      }
      console.warn("Warning: could not verify checksum —", checksumErr.message);
    }

    if (isWindows) {
      execFileSync("powershell", [
        "-Command",
        `Expand-Archive -Path '${archivePath}' -DestinationPath '${tmpDir}'`,
      ], { stdio: "ignore" });
    } else {
      execFileSync("tar", ["-xzf", archivePath, "-C", tmpDir], {
        stdio: "ignore",
      });
    }

    const binaryName = NAME + (isWindows ? ".exe" : "");
    const extractedBinary = path.join(tmpDir, binaryName);

    fs.copyFileSync(extractedBinary, dest);
    fs.chmodSync(dest, 0o755);
    console.log(`${NAME} v${VERSION} installed successfully`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

try {
  install();
} catch (err) {
  console.error(`Failed to install ${NAME}:`, err.message);
  console.error(
    `\nYou can manually download the binary from:\n` +
      `  https://github.com/${REPO}/releases\n`
  );
  process.exit(1);
}

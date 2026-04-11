#!/usr/bin/env node
const { execFileSync } = require("child_process");
const path = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = path.join(__dirname, "..", "bin", "tcl" + ext);

try {
  execFileSync(bin, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  if (e.code === "ENOENT") {
    console.error("Binary not found. Run 'npm install -g @tcl-eaglelab-01/smart-home-cli' to reinstall.");
  }
  process.exit(e.status || 1);
}

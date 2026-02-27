#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');
const os = require('os');

const REPO = 'JackUait/ghost-tab';

// Allow overrides for testing
const home = process.env.HOME || os.homedir();
const installDir = process.env.GHOST_TAB_INSTALL_DIR
  || path.join(home, '.local', 'share', 'ghost-tab');
const tuiBinDir = path.join(home, '.local', 'bin');
const tuiBinPath = path.join(tuiBinDir, 'ghost-tab-tui');

// Package root (where npm extracted us)
const pkgRoot = path.resolve(__dirname, '..');

function main() {
  // Platform check
  const platform = process.env.GHOST_TAB_MOCK_PLATFORM || process.platform;
  if (platform !== 'darwin') {
    process.stderr.write(`Error: ghost-tab only supports macOS (detected: ${platform})\n`);
    process.exit(1);
  }

  const version = fs.readFileSync(path.join(pkgRoot, 'VERSION'), 'utf8').trim();

  // Check if already installed at correct version
  const versionMarker = path.join(installDir, '.version');
  let installedVersion = '';
  try {
    installedVersion = fs.readFileSync(versionMarker, 'utf8').trim();
  } catch (_) {
    // Not installed yet
  }

  if (installedVersion === version) {
    process.stdout.write(`ghost-tab ${version} already up to date\n`);
  } else {
    // Copy bash distribution to install dir
    process.stdout.write(`Installing ghost-tab ${version} to ${installDir}...\n`);
    copyDistribution(pkgRoot, installDir);
    fs.writeFileSync(versionMarker, version + '\n');
    process.stdout.write(`Installed ghost-tab ${version}\n`);
  }

  // Download TUI binary if needed
  if (!process.env.GHOST_TAB_SKIP_TUI_DOWNLOAD) {
    ensureTuiBinary(version);
  }

  // Exec the bash installer
  if (!process.env.GHOST_TAB_SKIP_EXEC) {
    const installer = path.join(installDir, 'bin', 'ghost-tab');
    const args = process.argv.slice(2);
    try {
      execFileSync('bash', [installer, ...args], { stdio: 'inherit' });
    } catch (err) {
      process.exit(err.status || 1);
    }
  }
}

// Recursively copy the bash distribution files.
function copyDistribution(src, dest) {
  const entries = [
    'bin/ghost-tab',
    'lib',
    'templates',
    'ghostty',
    'terminals',
    'wrapper.sh',
    'VERSION',
  ];

  for (const entry of entries) {
    const srcPath = path.join(src, entry);
    if (!fs.existsSync(srcPath)) continue;
    const destPath = path.join(dest, entry);
    copyRecursive(srcPath, destPath);
  }
}

function copyRecursive(src, dest) {
  const stat = fs.statSync(src);
  if (stat.isDirectory()) {
    fs.mkdirSync(dest, { recursive: true });
    for (const child of fs.readdirSync(src)) {
      copyRecursive(path.join(src, child), path.join(dest, child));
    }
  } else {
    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.copyFileSync(src, dest);
    // Preserve executable bit
    if (stat.mode & 0o111) {
      fs.chmodSync(dest, stat.mode);
    }
  }
}

// Download the TUI binary from GitHub Releases if missing or wrong version.
function ensureTuiBinary(version) {
  // Check if existing binary matches version
  try {
    const out = execFileSync(tuiBinPath, ['--version'], { encoding: 'utf8' });
    const installed = out.replace(/.*version\s*/, '').trim();
    if (installed === version) {
      process.stdout.write(`ghost-tab-tui ${version} already up to date\n`);
      return;
    }
    process.stdout.write(`Updating ghost-tab-tui (${installed} -> ${version})...\n`);
  } catch (_) {
    process.stdout.write(`Downloading ghost-tab-tui ${version}...\n`);
  }

  const arch = process.arch === 'x64' ? 'amd64' : process.arch;
  const url = `https://github.com/${REPO}/releases/download/v${version}/ghost-tab-tui-darwin-${arch}`;

  fs.mkdirSync(tuiBinDir, { recursive: true });
  downloadFile(url, tuiBinPath);
  fs.chmodSync(tuiBinPath, 0o755);
  process.stdout.write(`ghost-tab-tui ${version} installed\n`);
}

// Synchronous HTTPS download with redirect following.
function downloadFile(url, dest) {
  execFileSync('curl', ['-fsSL', '-o', dest, url], {
    stdio: ['pipe', 'pipe', 'pipe'],
  });
}

main();

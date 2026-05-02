#!/usr/bin/env node
// Wrapper that invokes the platform-specific optiqor binary fetched by
// postinstall.js. Stays small and dependency-free on purpose.

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const binaryName = process.platform === 'win32' ? 'optiqor.exe' : 'optiqor';
const binaryPath = path.join(__dirname, '..', 'vendor', binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error('optiqor: binary not found at', binaryPath);
  console.error('optiqor: try `npm install -g @optiqor/cli` again, or build from source:');
  console.error('  go install github.com/optiqor/optiqor-cli/cmd/optiqor@latest');
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: true,
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 1);
  }
});

child.on('error', (err) => {
  console.error('optiqor: failed to launch binary:', err.message);
  process.exit(1);
});

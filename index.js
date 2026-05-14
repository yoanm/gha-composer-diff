import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

let binPath;

try {
    binPath = chooseBinary();
} catch (err) {
    console.log(`::error::Unable to pick a wrapper: ${err}`)
    process.exit(1)
}

if (false === path.existsSync(binPath)) {
    console.log(`::error::Binary wrapper not found: ${path.basename(binPath)}`)
    process.exit(2)
}

const { status, error } = spawnSync(chooseBinary(), { stdio: 'inherit' })
if (status !== 0) {
    if (undefined !== error) {
        console.log(`::error::Binary wrapper exited with code ${status} and errors: ${error}`)
    } else {
        console.log(`::error::Binary wrapper non-zero exit code: ${status}`)
    }
    process.exit(3)
}

console.log("::debug::Binary wrapper successfully executed")
process.exit(0)

function chooseBinary() {
    let platform = os.platform();
    switch (platform) {
        case 'win32':
            platform = 'windows';
            break;
        case 'darwin':
        case 'linux':
            break;
        default:
            throw new Error(`Unsupported architecture: ${platform}`);
    }
    let arch = os.arch();
    switch (arch) {
        case 'x64':
            arch = 'amd64';
            break;
        case 'arm64':
            break;
        default:
            throw new Error(`Unsupported architecture: ${arch}`);
    }

    return path.join(__dirname, 'bin', platform + '-' + arch + '-wrapper')
}

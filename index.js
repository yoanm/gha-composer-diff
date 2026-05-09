const os = require('os');
const path = require('path');
const { spawnSync } = require('child_process');

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

    return path.join('bin', platform + '-' + arch + '-wrapper')
}

const res = spawnSync(path.join(__dirname, chooseBinary()), { stdio: 'inherit' })

console.log(JSON.stringify(res))

if (res.status !== 0) {
    throw new Error(`Failed to execute binary, exit code ${res.status} with error: ${res.error}`);
}

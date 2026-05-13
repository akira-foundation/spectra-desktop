#!/usr/bin/env node

import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..')

const rootPkgPath = path.join(root, 'package.json')
const frontendPkgPath = path.join(root, 'frontend/package.json')
const wailsJsonPath = path.join(root, 'wails.json')
const versionGoPath = path.join(root, 'internal/version/version.go')

const rootPkg = JSON.parse(fs.readFileSync(rootPkgPath, 'utf8'))
const version = rootPkg.version

let changed = false

// frontend/package.json
const frontendPkg = JSON.parse(fs.readFileSync(frontendPkgPath, 'utf8'))
if (frontendPkg.version !== version) {
  frontendPkg.version = version
  fs.writeFileSync(frontendPkgPath, JSON.stringify(frontendPkg, null, 2) + '\n')
  changed = true
}

// wails.json info.productVersion
const wailsJson = JSON.parse(fs.readFileSync(wailsJsonPath, 'utf8'))
wailsJson.info = wailsJson.info || {}
if (wailsJson.info.productVersion !== version) {
  wailsJson.info.productVersion = version
  fs.writeFileSync(wailsJsonPath, JSON.stringify(wailsJson, null, 2) + '\n')
  changed = true
}

// internal/version/version.go
fs.mkdirSync(path.dirname(versionGoPath), { recursive: true })
const versionGo = `package version

// Version is the application version.
// Overridden at build time via -ldflags "-X spectra-desktop/internal/version.Version=x.y.z".
var Version = "${version}"
`
const existing = fs.existsSync(versionGoPath) ? fs.readFileSync(versionGoPath, 'utf8') : ''
if (existing !== versionGo) {
  fs.writeFileSync(versionGoPath, versionGo)
  changed = true
}

console.log(changed ? `Synced version to ${version}` : `Version already synced (${version})`)

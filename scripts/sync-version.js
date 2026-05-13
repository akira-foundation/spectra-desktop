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

// internal/version/version.go intentionally keeps the default "dev" so that
// any non-CI build (wails dev, local `wails build` without ldflags) is
// recognised as dev. CI overrides via -ldflags at link time.
// Touch nothing here.
void versionGoPath

console.log(changed ? `Synced version to ${version}` : `Version already synced (${version})`)

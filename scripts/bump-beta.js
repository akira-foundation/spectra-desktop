#!/usr/bin/env node

import { execSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..')
const pkgPath = path.join(root, 'package.json')

const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'))
const current = pkg.version

const match = current.match(/^(\d+)\.(\d+)\.(\d+)(?:-beta\.(\d+))?$/)
if (!match) {
  die(`unsupported version ${current}; expected X.Y.Z or X.Y.Z-beta.N`)
}

const [, maj, min, patch, betaStr] = match
const next = betaStr
  ? `${maj}.${min}.${patch}-beta.${Number(betaStr) + 1}`
  : `${maj}.${min}.${patch}-beta.1`

run('git diff --quiet && git diff --cached --quiet', 'working tree not clean — commit or stash first')

const tag = `v${next}`
checkTagFree(tag)

pkg.version = next
fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + '\n')

run('node scripts/sync-version.js')
run('git add package.json frontend/package.json wails.json')
run(`git commit -m "chore(release): ${next}"`)
run('git push')
run(`git tag ${tag}`)
run(`git push origin ${tag}`)

console.log('')
console.log(`released ${next}`)
console.log(`tag      ${tag}`)
console.log('CI:      https://github.com/akira-foundation/spectra-desktop/actions')

function run(cmd, errMsg) {
  try {
    execSync(cmd, { stdio: 'inherit', cwd: root })
  } catch (err) {
    die(errMsg ?? `failed: ${cmd}`)
  }
}

function checkTagFree(tag) {
  try {
    execSync(`git rev-parse -q --verify refs/tags/${tag}`, { cwd: root, stdio: 'ignore' })
    die(`tag ${tag} already exists locally`)
  } catch {
    // good — tag does not exist
  }
  try {
    const out = execSync(`git ls-remote --tags origin ${tag}`, { cwd: root, encoding: 'utf8' })
    if (out.trim()) die(`tag ${tag} already exists on remote`)
  } catch {
    // network failure shouldn't block; ignore
  }
}

function die(msg) {
  console.error(`bump-beta: ${msg}`)
  process.exit(1)
}

# Builds site/ : web/* + main.wasm + wasm_exec.js + graph (web/graph and graph.json).
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $env:Path += ";C:\Program Files\Go\bin"
}

function Invoke-Go {
    & go @args
    if ($LASTEXITCODE -ne 0) { throw "go $($args -join ' ') failed" }
}

$site = Join-Path $PSScriptRoot 'site'
if (Test-Path $site) { Remove-Item $site -Recurse -Force }
New-Item -ItemType Directory -Path $site | Out-Null

# Page files, without the graph folder (copied separately below).
Get-ChildItem -Path 'web' -Exclude 'graph' | Copy-Item -Destination $site -Recurse

# WebAssembly build.
$env:GOOS = 'js'
$env:GOARCH = 'wasm'
try {
    Invoke-Go build -o (Join-Path $site 'main.wasm') ./cmd/wasm
} finally {
    Remove-Item Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
}

# The runtime shim lives in lib/wasm on recent Go, misc/wasm on older ones.
$goroot = (& go env GOROOT).Trim()
$shim = @('lib\wasm\wasm_exec.js', 'misc\wasm\wasm_exec.js') |
    ForEach-Object { Join-Path $goroot $_ } |
    Where-Object { Test-Path $_ } |
    Select-Object -First 1
if (-not $shim) { throw "wasm_exec.js not found under $goroot" }
Copy-Item $shim (Join-Path $site 'wasm_exec.js')

# Code graph page and data.
$graphSite = Join-Path $site 'graph'
if (Test-Path 'web/graph') {
    Copy-Item 'web/graph' $graphSite -Recurse
}
if (Test-Path 'cmd/codegraph') {
    New-Item -ItemType Directory -Path $graphSite -Force | Out-Null
    Invoke-Go run ./cmd/codegraph -o (Join-Path $graphSite 'graph.json')
}

Write-Host "site/ ready: $((Get-ChildItem $site -Recurse -File).Count) files"

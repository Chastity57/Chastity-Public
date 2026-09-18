# 全栈开发1 · 一键启动
# 启动桥接服务：它会编译并运行原后端 main.go（源码不改），
# 同时托管 web/ 里的登录页面，并自动打开浏览器。

$ErrorActionPreference = 'Stop'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host '没有找到 go 命令，请先安装 Go 并把它加入 PATH。' -ForegroundColor Red
    Read-Host '按回车键退出'
    exit 1
}

Set-Location -LiteralPath (Join-Path $PSScriptRoot 'server')

Write-Host '正在启动桥接服务…' -ForegroundColor Cyan
Write-Host '页面地址： http://127.0.0.1:8080/' -ForegroundColor Cyan
Write-Host '关闭后端： 在本窗口按 Ctrl+C，或直接点网页上的“3 退出”。' -ForegroundColor DarkGray
Write-Host ''

go run main.go

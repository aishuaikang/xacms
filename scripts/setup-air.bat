@echo off
REM 为Windows设置air配置

echo 正在为Windows设置air配置...

(
echo root = "."
echo testdata_dir = "testdata"
echo tmp_dir = "tmp"
echo.
echo [build]
echo   args_bin = []
echo   bin = "./main.exe"
echo   cmd = "make build"
echo   delay = 1000
echo   exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules"]
echo   exclude_file = []
echo   exclude_regex = ["_test.go"]
echo   exclude_unchanged = false
echo   follow_symlink = false
echo   full_bin = ""
echo   include_dir = []
echo   include_ext = ["go", "tpl", "tmpl", "html"]
echo   include_file = []
echo   kill_delay = "0s"
echo   log = "build-errors.log"
echo   poll = false
echo   poll_interval = 0
echo   post_cmd = []
echo   pre_cmd = []
echo   rerun = false
echo   rerun_delay = 500
echo   send_interrupt = false
echo   stop_on_error = false
echo.
echo [color]
echo   app = ""
echo   build = "yellow"
echo   main = "magenta"
echo   runner = "green"
echo   watcher = "cyan"
echo.
echo [log]
echo   main_only = false
echo   time = false
echo.
echo [misc]
echo   clean_on_exit = false
echo.
echo [screen]
echo   clear_on_rebuild = false
echo   keep_scroll = true
) > .air.toml

echo Windows的Air配置已更新，二进制文件: ./main.exe

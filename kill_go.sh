ps aux | grep go-build | grep -v grep | cut -d ' ' -f 2 | xargs -i kill {}

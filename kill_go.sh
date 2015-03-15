ps aux | grep go-build | grep -v grep | cut -d ' ' -f 3 | xargs -i kill {}

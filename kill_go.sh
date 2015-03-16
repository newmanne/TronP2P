ps aux | grep go-build | grep -v grep | awk '{print $2}' | xargs kill

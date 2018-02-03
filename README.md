# oodle

[![Build Status](https://travis-ci.org/godwhoa/oodle.svg?branch=master)](https://travis-ci.org/godwhoa/oodle)

> **oodle** is a simple irc bot for oodnet. It is also a rewrite of ezbot framework 

## Timeline
[x] Base framework<br>
[x] Config. and logging<br>
[x] Make it robust (reconnect and recovery)<br>
[x] Core commands<br>
[ ] oodnet integration<br>
[ ] More polish (.help <command> etc.)<br>
[ ] Simplify design if possible<br>

## Usage
```bash
# Download latest build
curl -s https://api.github.com/repos/godwhoa/oodle/releases/latest | grep browser_download_url | cut -d '"' -f 4 | xargs -L 1 wget
# Edit config
vim config.toml
# Make it executable
chmod +x oodle_linux_static
# Run!
./oodle_linux_static
```
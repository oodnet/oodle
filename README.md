# oodle

[![Build Status](https://travis-ci.org/oodnet/oodle.svg?branch=master)](https://travis-ci.org/oodnet/oodle)

> **oodle** is a simple irc bot for oodnet. It is also a rewrite of ezbot framework 

## Features 
- Title scraper
- Reminder
- Tell
- Seen
- Built-in upgrade mechanism
- Github notifier (alternative to github's IRC service)
- Webhook
- Urban
- Wiki
- Find and Replace with sed-like syntax
- Custom Commands

## Supported platforms
- Linux AMD64
- Linux ARMv7

## Usage
```bash
# Download latest build
# Note: remove -v from grep if you are on ARM
curl -s https://api.github.com/repos/oodnet/oodle/releases/latest | grep browser_download_url | cut -d '"' -f 4 | grep -v arm | xargs -L 1 wget -O oodle
# Edit config
vim config.toml
# Make it executable
chmod +x oodle_linux
# Run!
./oodle_linux
```
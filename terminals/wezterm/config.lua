local wezterm = require 'wezterm'
local config = wezterm.config_builder()
config.default_prog = { '~/.config/wisp-deck/wrapper.sh' }
return config

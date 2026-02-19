local wezterm = require 'wezterm'
local config = wezterm.config_builder()
config.default_prog = { '~/.config/ghost-tab/wrapper.sh' }
return config

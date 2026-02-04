class GhostTab < Formula
  desc "Ghostty + tmux wrapper for AI-assisted development with Claude Code"
  homepage "https://github.com/JackUait/ghost-tab"
  url "https://github.com/JackUait/ghost-tab/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "3efaa2c32ebad02131167b9a41e38835d22215971d3dc2f33fa50ef1d5be6416"
  license "MIT"

  depends_on "tmux"
  depends_on "lazygit"
  depends_on "broot"
  depends_on "claude-code"
  depends_on "node@22"
  depends_on :macos

  def install
    bin.install "bin/ghost-tab"
    (share/"ghost-tab").install "ghostty"
  end

  def caveats
    <<~EOS
      Ghostty terminal is required but not installed automatically.
      Install it with:
        brew install --cask ghostty

      Then run `ghost-tab` to set up your environment.
    EOS
  end

  test do
    assert_match "macOS", shell_output("#{bin}/ghost-tab --help 2>&1", 1)
  end
end

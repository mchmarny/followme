class Daprme < Formula
  desc "Utility to monitor Twitter followers"
  homepage "https://thingz.io"
  url "https://github.com/mchmarny/followme/releases/download/v0.2.12/followme"
  sha256 "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 "
  license "MIT"

  def install
    bin.install "followme" => "followme"
  end

  test do
    system "#{bin}/followme", "--version"
  end
end
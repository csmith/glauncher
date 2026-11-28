pkgname=glauncher
pkgver=0
pkgrel=1
pkgdesc='Search-driven application launcher'
arch=('x86_64')
url='https://git.yak-wall.ts.net/public/glauncher'
license=('MIT')
depends=('glibc' 'libglvnd' 'libx11' 'libxcursor' 'libxfixes' 'libxkbcommon' 'libxkbcommon-x11' 'wayland' 'xdg-utils')
makedepends=('git' 'go' 'vulkan-headers')
optdepends=('wl-clipboard: copying results on Wayland'
            'xclip: copying results on X11'
            'xsel: copying results on X11')

pkgver() {
  cd "$startdir"
  if tag=$(git describe --long --tags --abbrev=7 2>/dev/null); then
    sed 's/^v//;s/\([^-]*-g\)/r\1/;s/-/./g' <<< "$tag"
  else
    printf '0.r%s.g%s' "$(git rev-list --count HEAD)" "$(git rev-parse --short=7 HEAD)"
  fi
}

build() {
  cd "$startdir"
  export CGO_CPPFLAGS="$CPPFLAGS" CGO_CFLAGS="$CFLAGS" CGO_CXXFLAGS="$CXXFLAGS" CGO_LDFLAGS="$LDFLAGS"
  export GOFLAGS='-buildmode=pie -trimpath -mod=readonly -modcacherw'
  make build
}

package() {
  cd "$startdir"
  install -Dm755 glauncher "$pkgdir/usr/bin/glauncher"
  install -Dm644 assets/com.chameth.glauncher.desktop "$pkgdir/usr/share/applications/com.chameth.glauncher.desktop"
  install -Dm644 assets/com.chameth.glauncher.svg "$pkgdir/usr/share/icons/hicolor/scalable/apps/com.chameth.glauncher.svg"
  install -Dm644 LICENCE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}

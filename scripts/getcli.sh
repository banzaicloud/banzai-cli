#!/bin/sh

set -e

usage() {
    echo
    echo "Install latest release of banzai-cli. https://github.com/banzaicloud/banzai-cli"
    echo
    if have curl; then
        echo "Usage: curl https://getpipeline.sh/cli | sh [-s -- auto|deb|rpm|brew|tar|go|kubectl]"
    elif have wget; then
        echo "Usage: wget -O- https://getpipeline.sh/cli | sh [-s -- auto|deb|rpm|brew|tar|go|kubectl]"
    else
        echo "Usage: $1 [auto|deb|rpm|brew|tar|go|kubectl]"
    fi
}

main() {
    cmd="$1"
    shift
    case "$*" in
        guess-only)
            guess_installer ;;
        "")
            if have banzai; then
                echo "banzai-cli is already installed. Run with 'auto' argument to install/upgrade anyway" >&2
                usage $cmd >&2
                exit 1
            fi
            install_`guess_installer` ;;
        auto)
            install_`guess_installer` ;;
        deb)
            install_deb ;;
        rpm)
            install_rpm ;;
        tar|tgz|tar.gz)
            install_tar ;;
        brew)
            install_brew ;;
        go)
            install_go ;;
        kubectl)  # install kubectl
            install_kubectl ;;
        *)
            usage $cmd >&2
            exit 1 
    esac
}

install_brew() {
    brew install banzaicloud/tap/banzai-cli
}

install_rpm() {
    rpm -i "https://banzaicloud.com/downloads/banzai-cli/latest?format=rpm"
}

install_deb() {
    tmp=`tmp`
    trap "rm -r $tmp" EXIT
    download "https://banzaicloud.com/downloads/banzai-cli/latest?format=deb" $tmp/banzai-cli.deb 
    sudo dpkg -i $tmp/banzai-cli.deb
}

install_tar() {
    tmp=`tmp`
    trap "rm -r $tmp" EXIT
    download "https://banzaicloud.com/downloads/banzai-cli/latest?format=tgz&os=`os`" | tar xz -C $tmp
    path=`path`
    if [ -w $path ]; then
        install $tmp/banzai $path
    else
        sudo install $tmp/banzai $path
    fi
}

install_go() {
    go get github.com/banzaicloud/banzai-cli/cmd/banzai
    sudo install -m 755 ${GOPATH:-~/go}/bin/banzai "`path`/banzai"
}

install_kubectl() {
    version=`download "https://storage.googleapis.com/kubernetes-release/release/stable.txt" 2>/dev/null`
    tmp=`tmp`
    trap "rm -r $tmp" EXIT
    download "https://storage.googleapis.com/kubernetes-release/release/${version}/bin/`os`/amd64/kubectl" $tmp/kubectl
    sudo install $tmp/kubectl "`path`/kubectl"
}

have() {
    type "$@" >/dev/null 2>&1
}

os() {
    case "`uname`" in
        Darwin) echo darwin ;;
        Linux) echo linux ;;
        *) echo Unsupported OS. >&2
            exit 1
    esac
}

if [ "`whoami`" = root ]; then
    sudo() {
        "$@"
    }
fi

download() {
    if have wget; then
            wget -O "${2:--}" "$1"
    elif have curl; then
            curl -L -o "${2:--}" "$1"
    else
        echo Neither wget, nor curl is available in the system PATH.
        exit 1
    fi
}

path() {
    if echo "$PATH" | tr : '\n' | grep -qs ^/usr/local/bin$; then
        echo /usr/local/bin
    else
        echo /usr/bin
    fi
}

tmp() {
    if have mktemp; then
        dir="`mktemp -d`"
    else
        dir=/tmp/$USER.$$
        mkdir -p $dir
    fi
    echo $dir
}

guess_installer() {
    case "`uname` `uname -m`" in
        Darwin*)
            if have brew; then
                echo brew
            else
                echo tgz
            fi
            ;;
        "Linux x86_64")
            [ ! -e /etc/os-release ] || . /etc/os-release
            case "$ID" in
                centos|rhel|fedora|sles|*suse*)
                    echo rpm
                    ;;
                debian|ubuntu)
                    echo deb
                    ;;
                *)
                    if have rpm; then
                        if have dpkg; then
                            echo "Can't decide between deb and rpm, falling back to tarball" >&2
                            echo tar
                        else
                            echo rpm
                        fi
                    else
                        if have dpkg; then
                            echo deb
                        else
                            echo tar
                        fi
                    fi
            esac
            ;;
        *)
            if have go; then
                echo go
            else
                echo Unsupported operating system. Try installing go to build from source. >&2
                exit 1
            fi
    esac
}

main "$0" "$@"

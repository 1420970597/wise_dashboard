#!/bin/sh

NZ_BASE_PATH="/opt/nezha"
NZ_AGENT_PATH="${NZ_BASE_PATH}/agent"

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
plain='\033[0m'

err() {
    printf "${red}%s${plain}\n" "$*" >&2
}

success() {
    printf "${green}%s${plain}\n" "$*"
}

info() {
    printf "${yellow}%s${plain}\n" "$*"
}

sudo() {
    myEUID=$(id -ru)
    if [ "$myEUID" -ne 0 ]; then
        if command -v sudo > /dev/null 2>&1; then
            command sudo "$@"
        else
            err "ERROR: sudo is not installed on the system, the action cannot be proceeded."
            exit 1
        fi
    else
        "$@"
    fi
}

deps_check() {
    local deps="curl grep"
    local _err=0
    local missing=""

    for dep in $deps; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            _err=1
            missing="${missing} $dep"
        fi
    done

    if [ "$_err" -ne 0 ]; then
        err "Missing dependencies:$missing. Please install them and try again."
        exit 1
    fi
}

env_check() {
    mach=$(uname -m)
    case "$mach" in
        amd64|x86_64)
            os_arch="amd64"
            ;;
        i386|i686)
            os_arch="386"
            ;;
        aarch64|arm64)
            os_arch="arm64"
            ;;
        *arm*)
            os_arch="arm"
            ;;
        s390x)
            os_arch="s390x"
            ;;
        riscv64)
            os_arch="riscv64"
            ;;
        mips)
            os_arch="mips"
            ;;
        mipsel|mipsle)
            os_arch="mipsle"
            ;;
        *)
            err "Unknown architecture: $mach"
            exit 1
            ;;
    esac

    system=$(uname)
    case "$system" in
        *Linux*)
            os="linux"
            ;;
        *Darwin*)
            os="darwin"
            ;;
        *FreeBSD*)
            os="freebsd"
            ;;
        *)
            err "Unknown OS: $system"
            exit 1
            ;;
    esac
}

init() {
    deps_check
    env_check
}

install() {
    echo "Installing..."

    # 检查必需的环境变量
    if [ -z "$NZ_SERVER" ]; then
        err "NZ_SERVER should not be empty"
        exit 1
    fi

    if [ -z "$NZ_CLIENT_SECRET" ]; then
        err "NZ_CLIENT_SECRET should not be empty"
        exit 1
    fi

    # 构建下载 URL
    if [ "$NZ_TLS" = "true" ]; then
        PROTOCOL="https"
    else
        PROTOCOL="http"
    fi

    NZ_AGENT_URL="${PROTOCOL}://${NZ_SERVER}/nezha-agent"

    # 下载 agent
    info "Downloading agent from $NZ_AGENT_URL"

    if command -v wget >/dev/null 2>&1; then
        _cmd="wget --timeout=60 -O /tmp/nezha-agent \"$NZ_AGENT_URL\" >/dev/null 2>&1"
    elif command -v curl >/dev/null 2>&1; then
        _cmd="curl --max-time 60 -fsSL \"$NZ_AGENT_URL\" -o /tmp/nezha-agent >/dev/null 2>&1"
    else
        err "Neither wget nor curl is available"
        exit 1
    fi

    if ! eval "$_cmd"; then
        err "Download nezha-agent failed, check your network connectivity"
        exit 1
    fi

    # 创建目录
    sudo mkdir -p $NZ_AGENT_PATH

    # 移动并设置权限
    sudo mv /tmp/nezha-agent $NZ_AGENT_PATH/nezha-agent
    sudo chmod +x $NZ_AGENT_PATH/nezha-agent

    # 创建配置文件路径
    path="$NZ_AGENT_PATH/config.yml"
    if [ -f "$path" ]; then
        random=$(LC_ALL=C tr -dc a-z0-9 </dev/urandom | head -c 5)
        path=$(printf "%s" "$NZ_AGENT_PATH/config-$random.yml")
    fi

    # 设置环境变量
    env="NZ_UUID=$NZ_UUID NZ_SERVER=$NZ_SERVER NZ_CLIENT_SECRET=$NZ_CLIENT_SECRET NZ_TLS=$NZ_TLS NZ_DISABLE_AUTO_UPDATE=$NZ_DISABLE_AUTO_UPDATE NZ_DISABLE_FORCE_UPDATE=$DISABLE_FORCE_UPDATE NZ_DISABLE_COMMAND_EXECUTE=$NZ_DISABLE_COMMAND_EXECUTE NZ_SKIP_CONNECTION_COUNT=$NZ_SKIP_CONNECTION_COUNT"

    # 先卸载旧服务(如果存在)
    sudo "${NZ_AGENT_PATH}"/nezha-agent service -c "$path" uninstall >/dev/null 2>&1

    # 安装服务
    info "Installing service..."
    _cmd="sudo env $env $NZ_AGENT_PATH/nezha-agent service -c $path install"
    if ! eval "$_cmd"; then
        err "Install nezha-agent service failed"
        sudo "${NZ_AGENT_PATH}"/nezha-agent service -c "$path" uninstall >/dev/null 2>&1
        exit 1
    fi

    success "nezha-agent successfully installed"
    info "Service status: systemctl status nezha-agent"
    info "View logs: journalctl -u nezha-agent -f"
}

uninstall() {
    find "$NZ_AGENT_PATH" -type f -name "*config*.yml" | while read -r file; do
        sudo "$NZ_AGENT_PATH/nezha-agent" service -c "$file" uninstall
        sudo rm "$file"
    done
    sudo rm -rf "$NZ_AGENT_PATH"
    info "Uninstallation completed."
}

if [ "$1" = "uninstall" ]; then
    uninstall
    exit
fi

init
install

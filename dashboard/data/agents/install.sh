#!/bin/sh

NZ_BASE_PATH="/opt/nezha"
NZ_AGENT_PATH="${NZ_BASE_PATH}/agent"
NZ_CONFIG_PATH="$NZ_AGENT_PATH/config.yml"

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

    # 设置协议
    if [ "$NZ_TLS" = "true" ]; then
        PROTOCOL="https"
    else
        PROTOCOL="http"
    fi
}

# 检测现有安装
detect_existing_installation() {
    if [ -f "$NZ_AGENT_PATH/nezha-agent" ]; then
        return 0  # 存在安装
    fi
    if systemctl is-enabled --quiet nezha-agent 2>/dev/null; then
        return 0  # 服务已注册
    fi
    return 1  # 不存在
}

# 下载 agent 二进制文件
download_agent() {
    NZ_AGENT_URL="${PROTOCOL}://${NZ_SERVER}/nezha-agent"

    info "Downloading agent from $NZ_AGENT_URL (this may take a while...)"

    download_success=0
    if command -v wget >/dev/null 2>&1; then
        if wget --timeout=180 --tries=3 --show-progress -O /tmp/nezha-agent "$NZ_AGENT_URL" 2>&1; then
            download_success=1
        fi
    elif command -v curl >/dev/null 2>&1; then
        if curl --max-time 180 --retry 3 --retry-delay 2 -fSL "$NZ_AGENT_URL" -o /tmp/nezha-agent 2>&1; then
            download_success=1
        fi
    else
        err "Neither wget nor curl is available"
        exit 1
    fi

    if [ $download_success -eq 0 ]; then
        err "Download nezha-agent failed, check your network connectivity"
        err "You can try to download manually: curl -L $NZ_AGENT_URL -o /tmp/nezha-agent"
        exit 1
    fi

    # 创建目录
    sudo mkdir -p $NZ_AGENT_PATH

    # 移动并设置权限
    sudo mv /tmp/nezha-agent $NZ_AGENT_PATH/nezha-agent
    sudo chmod +x $NZ_AGENT_PATH/nezha-agent
}

# 配置审计功能
configure_audit() {
    local config_file="$1"

    # 从 NZ_SERVER 提取 host 和 port
    local host=$(echo "$NZ_SERVER" | cut -d':' -f1)
    local port=$(echo "$NZ_SERVER" | cut -d':' -f2)

    # 如果端口与 host 相同（没有指定端口），使用默认端口
    if [ "$port" = "$host" ]; then
        port="8008"
    fi

    # 构建 Dashboard URL
    local dashboard_url="${PROTOCOL}://${host}:${port}"

    # 追加审计配置到配置文件
    cat >> "$config_file" << EOF

# 终端审计配置（自动生成）
audit_enabled: true
audit_dashboard_url: "${dashboard_url}"
audit_token: "${NZ_CLIENT_SECRET}"
EOF

    info "Terminal audit configured: $dashboard_url"
}

# 检查并补充审计配置
ensure_audit_config() {
    local config_file="$1"

    if [ ! -f "$config_file" ]; then
        return 1
    fi

    # 检查是否已有审计配置
    if grep -q "audit_enabled" "$config_file" 2>/dev/null; then
        info "Audit config already exists, skipping..."
        return 0
    fi

    # 补充审计配置
    info "Adding missing audit config..."
    configure_audit "$config_file"
    return 0
}

# 升级现有安装
upgrade() {
    info "Detected existing installation, upgrading..."

    # 检查必需的环境变量
    if [ -z "$NZ_SERVER" ]; then
        err "NZ_SERVER should not be empty"
        exit 1
    fi

    if [ -z "$NZ_CLIENT_SECRET" ]; then
        err "NZ_CLIENT_SECRET should not be empty"
        exit 1
    fi

    # 查找配置文件
    local config_file="$NZ_CONFIG_PATH"
    if [ ! -f "$config_file" ]; then
        # 尝试查找其他配置文件
        config_file=$(find "$NZ_AGENT_PATH" -type f -name "config*.yml" 2>/dev/null | head -1)
    fi

    # 备份现有配置
    if [ -n "$config_file" ] && [ -f "$config_file" ]; then
        info "Backing up config: $config_file"
        sudo cp "$config_file" "${config_file}.bak.$(date +%Y%m%d_%H%M%S)"
    fi

    # 停止服务
    info "Stopping service..."
    sudo systemctl stop nezha-agent 2>/dev/null || true

    # 备份旧二进制
    if [ -f "$NZ_AGENT_PATH/nezha-agent" ]; then
        sudo mv "$NZ_AGENT_PATH/nezha-agent" "$NZ_AGENT_PATH/nezha-agent.old"
    fi

    # 下载新二进制
    download_agent

    # 检查并补充审计配置
    if [ -n "$config_file" ] && [ -f "$config_file" ]; then
        ensure_audit_config "$config_file"
    fi

    # 重启服务
    info "Starting service..."
    sudo systemctl start nezha-agent

    # 检查服务状态
    sleep 2
    if systemctl is-active --quiet nezha-agent; then
        success "Agent upgraded successfully!"

        # 清理旧二进制
        sudo rm -f "$NZ_AGENT_PATH/nezha-agent.old"
    else
        err "Service failed to start, rolling back..."
        sudo mv "$NZ_AGENT_PATH/nezha-agent.old" "$NZ_AGENT_PATH/nezha-agent"
        sudo systemctl start nezha-agent
        exit 1
    fi

    info "Service status: systemctl status nezha-agent"
    info "View logs: journalctl -u nezha-agent -f"
}

# 全新安装
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

    # 下载 agent
    download_agent

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

    # 等待配置文件生成
    sleep 2

    # 配置审计功能
    if [ -f "$path" ]; then
        configure_audit "$path"
        # 重启服务以应用新配置
        info "Restarting service to apply audit config..."
        sudo systemctl restart nezha-agent
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

# 主入口
if [ "$1" = "uninstall" ]; then
    uninstall
    exit
fi

init

# 检测现有安装并决定操作
if detect_existing_installation; then
    # 检查是否强制重装
    if [ "$NZ_FORCE_REINSTALL" = "true" ]; then
        info "Force reinstall requested, removing existing installation..."
        uninstall
        install
    else
        # 默认升级
        upgrade
    fi
else
    # 全新安装
    install
fi

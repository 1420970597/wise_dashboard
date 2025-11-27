#!/bin/bash

#######################################
# Wise Dashboard 系统服务管理脚本
# 功能：安装、卸载、更新、状态检查
#######################################

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
SERVICE_NAME="wise-dashboard"
INSTALL_DIR="/opt/wise-dashboard"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 打印函数
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        error "请使用 root 权限运行此脚本"
        exit 1
    fi
}

# 检查 dashboard 可执行文件
check_dashboard_binary() {
    if [ ! -f "$SCRIPT_DIR/dashboard/dashboard" ]; then
        error "未找到 dashboard 可执行文件: $SCRIPT_DIR/dashboard/dashboard"
        error "请先编译项目: cd dashboard && go build -o dashboard ./cmd/dashboard"
        exit 1
    fi
}

# 创建 systemd 服务文件
create_service_file() {
    info "创建 systemd 服务文件..."

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Wise Dashboard Service
Documentation=https://github.com/nezhahq/nezha
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/dashboard
Restart=always
RestartSec=5s
StandardOutput=journal
StandardError=journal

# 安全选项
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${INSTALL_DIR}/data
ReadOnlyPaths=${INSTALL_DIR}

# 资源限制
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

    if [ $? -eq 0 ]; then
        success "服务文件创建成功: $SERVICE_FILE"
    else
        error "服务文件创建失败"
        exit 1
    fi
}

# 安装服务
install_service() {
    info "开始安装 Wise Dashboard 服务..."

    # 检查必要条件
    check_dashboard_binary

    # 停止现有服务（如果存在）
    if systemctl is-active --quiet $SERVICE_NAME; then
        warn "检测到服务正在运行，停止服务..."
        systemctl stop $SERVICE_NAME
    fi

    # 创建安装目录
    info "创建安装目录: $INSTALL_DIR"
    mkdir -p $INSTALL_DIR/data

    # 复制 dashboard 可执行文件
    info "复制 dashboard 可执行文件..."
    cp -f "$SCRIPT_DIR/dashboard/dashboard" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/dashboard"

    # 复制或保留 data 目录
    if [ -d "$SCRIPT_DIR/dashboard/data" ]; then
        info "复制配置和数据文件..."
        # 如果目标目录已经存在配置文件，备份
        if [ -f "$INSTALL_DIR/data/config.yaml" ]; then
            warn "检测到现有配置文件，创建备份..."
            cp "$INSTALL_DIR/data/config.yaml" "$INSTALL_DIR/data/config.yaml.bak.$(date +%Y%m%d_%H%M%S)"
        fi

        # 复制文件但不覆盖现有配置
        cp -n "$SCRIPT_DIR/dashboard/data/"* "$INSTALL_DIR/data/" 2>/dev/null || true

        # 复制 agents 目录（包含 agent 二进制和安装脚本）
        if [ -d "$SCRIPT_DIR/dashboard/data/agents" ]; then
            info "复制 agent 文件和安装脚本..."
            mkdir -p "$INSTALL_DIR/data/agents"
            cp -rf "$SCRIPT_DIR/dashboard/data/agents/"* "$INSTALL_DIR/data/agents/"
        fi

        # 复制 recordings 目录（如果存在）
        if [ -d "$SCRIPT_DIR/dashboard/data/recordings" ]; then
            mkdir -p "$INSTALL_DIR/data/recordings"
        fi
    fi

    # 设置权限
    info "设置文件权限..."
    chown -R root:root "$INSTALL_DIR"
    chmod -R 755 "$INSTALL_DIR"
    chmod 600 "$INSTALL_DIR/data/config.yaml" 2>/dev/null || true
    chmod 644 "$INSTALL_DIR/data/sqlite.db" 2>/dev/null || true

    # 创建服务文件
    create_service_file

    # 重新加载 systemd
    info "重新加载 systemd 配置..."
    systemctl daemon-reload

    # 启用服务
    info "启用服务开机自启..."
    systemctl enable $SERVICE_NAME

    # 启动服务
    info "启动服务..."
    systemctl start $SERVICE_NAME

    # 等待服务启动
    sleep 3

    # 检查服务状态
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo ""
        success "✓ Wise Dashboard 服务安装成功并正在运行！"
        echo ""
        info "安装位置: $INSTALL_DIR"
        info "服务文件: $SERVICE_FILE"
        echo ""
        info "服务管理命令："
        echo "  启动服务: systemctl start $SERVICE_NAME"
        echo "  停止服务: systemctl stop $SERVICE_NAME"
        echo "  重启服务: systemctl restart $SERVICE_NAME"
        echo "  查看状态: systemctl status $SERVICE_NAME"
        echo "  查看日志: journalctl -u $SERVICE_NAME -f"
        echo "  开机自启: systemctl enable $SERVICE_NAME"
        echo "  禁用自启: systemctl disable $SERVICE_NAME"
        echo ""
        info "当前服务状态："
        systemctl status $SERVICE_NAME --no-pager -l
    else
        error "✗ 服务启动失败，请查看日志："
        echo "  journalctl -u $SERVICE_NAME -n 50 --no-pager"
        echo ""
        journalctl -u $SERVICE_NAME -n 30 --no-pager
        exit 1
    fi
}

# 卸载服务
uninstall_service() {
    warn "开始卸载 Wise Dashboard 服务..."

    # 停止服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        info "停止服务..."
        systemctl stop $SERVICE_NAME
    fi

    # 禁用服务
    if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
        info "禁用服务..."
        systemctl disable $SERVICE_NAME
    fi

    # 删除服务文件
    if [ -f "$SERVICE_FILE" ]; then
        info "删除服务文件..."
        rm -f "$SERVICE_FILE"
    fi

    # 重新加载 systemd
    info "重新加载 systemd 配置..."
    systemctl daemon-reload
    systemctl reset-failed 2>/dev/null || true

    # 询问是否删除数据
    echo ""
    read -p "是否删除安装目录和数据？($INSTALL_DIR) [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "删除安装目录..."
        rm -rf "$INSTALL_DIR"
        success "✓ 服务和数据已完全卸载"
    else
        info "保留安装目录: $INSTALL_DIR"
        success "✓ 服务已卸载，数据已保留"
    fi
}

# 更新服务
update_service() {
    info "开始更新 Wise Dashboard 服务..."

    # 检查必要条件
    check_dashboard_binary

    # 停止服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        info "停止服务..."
        systemctl stop $SERVICE_NAME
    fi

    # 备份当前版本
    if [ -f "$INSTALL_DIR/dashboard" ]; then
        info "备份当前版本..."
        cp "$INSTALL_DIR/dashboard" "$INSTALL_DIR/dashboard.bak.$(date +%Y%m%d_%H%M%S)"
    fi

    # 复制新版本
    info "更新可执行文件..."
    cp -f "$SCRIPT_DIR/dashboard/dashboard" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/dashboard"

    # 启动服务
    info "启动服务..."
    systemctl start $SERVICE_NAME

    # 等待服务启动
    sleep 3

    # 检查服务状态
    if systemctl is-active --quiet $SERVICE_NAME; then
        success "✓ 服务更新成功并已启动"
        systemctl status $SERVICE_NAME --no-pager -l
    else
        error "✗ 服务启动失败，请查看日志"
        journalctl -u $SERVICE_NAME -n 30 --no-pager

        # 提示恢复备份
        warn "如需恢复旧版本："
        echo "  ls -lt $INSTALL_DIR/dashboard.bak.*"
        echo "  cp $INSTALL_DIR/dashboard.bak.XXXXXX $INSTALL_DIR/dashboard"
        echo "  systemctl start $SERVICE_NAME"
        exit 1
    fi
}

# 查看服务状态
show_status() {
    info "服务状态："
    systemctl status $SERVICE_NAME --no-pager -l

    echo ""
    info "最近日志："
    journalctl -u $SERVICE_NAME -n 20 --no-pager
}

# 显示帮助信息
show_help() {
    cat <<EOF
${BLUE}Wise Dashboard 系统服务管理脚本${NC}

用法: $0 [命令]

命令:
  install    安装并启动服务
  uninstall  停止并卸载服务
  update     更新服务（保留配置和数据）
  status     查看服务状态和日志
  help       显示此帮助信息

示例:
  $0 install     # 首次安装
  $0 update      # 更新版本
  $0 uninstall   # 卸载服务
  $0 status      # 查看状态

服务管理命令:
  systemctl start $SERVICE_NAME      # 启动服务
  systemctl stop $SERVICE_NAME       # 停止服务
  systemctl restart $SERVICE_NAME    # 重启服务
  systemctl status $SERVICE_NAME     # 查看状态
  journalctl -u $SERVICE_NAME -f     # 查看实时日志
EOF
}

# 主函数
main() {
    # 检查 root 权限（除了 help 命令）
    if [ "$1" != "help" ] && [ "$1" != "-h" ] && [ "$1" != "--help" ]; then
        check_root
    fi

    case "$1" in
        install)
            install_service
            ;;
        uninstall)
            uninstall_service
            ;;
        update)
            update_service
            ;;
        status)
            show_status
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"

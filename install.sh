#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息函数
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]; then
    error "请使用root权限运行此脚本"
    exit 1
fi

info "开始安装 Wise Dashboard 服务..."

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="/opt/wise-dashboard"
SERVICE_NAME="wise-dashboard"

# 检查dashboard可执行文件是否存在
if [ ! -f "$SCRIPT_DIR/dashboard/dashboard" ]; then
    error "未找到dashboard可执行文件，请先编译项目"
    error "运行: cd $SCRIPT_DIR/dashboard && go build -o dashboard ./cmd/dashboard"
    exit 1
fi

# 停止现有服务（如果存在）
if systemctl is-active --quiet $SERVICE_NAME; then
    info "停止现有的 $SERVICE_NAME 服务..."
    systemctl stop $SERVICE_NAME
fi

# 创建安装目录
info "创建安装目录: $INSTALL_DIR"
mkdir -p $INSTALL_DIR

# 复制dashboard可执行文件
info "复制dashboard可执行文件..."
cp -f $SCRIPT_DIR/dashboard/dashboard $INSTALL_DIR/

# 复制或创建data目录
if [ -d "$SCRIPT_DIR/dashboard/data" ]; then
    info "复制现有data目录..."
    cp -rf $SCRIPT_DIR/dashboard/data $INSTALL_DIR/
else
    info "创建新的data目录..."
    mkdir -p $INSTALL_DIR/data
fi

# 设置权限
info "设置文件权限..."
chmod +x $INSTALL_DIR/dashboard
chmod -R 755 $INSTALL_DIR

# 安装systemd服务文件
info "安装systemd服务文件..."
cp -f $SCRIPT_DIR/wise-dashboard.service /etc/systemd/system/

# 重新加载systemd配置
info "重新加载systemd配置..."
systemctl daemon-reload

# 启用并启动服务
info "启用并启动 $SERVICE_NAME 服务..."
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

# 等待服务启动
sleep 2

# 检查服务状态
if systemctl is-active --quiet $SERVICE_NAME; then
    info "✓ $SERVICE_NAME 服务安装成功并正在运行！"
    echo ""
    info "服务管理命令："
    echo "  启动服务: systemctl start $SERVICE_NAME"
    echo "  停止服务: systemctl stop $SERVICE_NAME"
    echo "  重启服务: systemctl restart $SERVICE_NAME"
    echo "  查看状态: systemctl status $SERVICE_NAME"
    echo "  查看日志: journalctl -u $SERVICE_NAME -f"
    echo ""
    info "服务状态："
    systemctl status $SERVICE_NAME --no-pager
else
    error "✗ 服务启动失败，请查看日志："
    echo "  journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi

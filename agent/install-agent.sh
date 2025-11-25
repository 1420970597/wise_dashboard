#!/bin/bash
# Nezha Agent 安装脚本（支持 AutoSSH）

set -e

DASHBOARD_HOST="156.227.233.135:8008"
AGENT_SECRET="Yo6ghXp61BZF4hJfArwSo3eUnVeQuXzg"
INSTALL_DIR="/opt/nezha/agent"

echo "=== Nezha Agent 安装脚本 ==="
echo "Dashboard: $DASHBOARD_HOST"
echo ""

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        AGENT_URL="http://156.227.233.135:9090/nezha-agent"
        ;;
    aarch64|arm64)
        echo "ARM64 架构，请手动编译 agent"
        exit 1
        ;;
    *)
        echo "不支持的架构: $ARCH"
        exit 1
        ;;
esac

# 创建目录
echo "创建安装目录..."
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 下载 agent
echo "下载 agent..."
if command -v wget &> /dev/null; then
    wget -O nezha-agent $AGENT_URL
elif command -v curl &> /dev/null; then
    curl -L -o nezha-agent $AGENT_URL
else
    echo "错误: 需要 wget 或 curl"
    exit 1
fi

# 赋予执行权限
chmod +x nezha-agent

# 创建 systemd 服务
echo "创建 systemd 服务..."
cat > /etc/systemd/system/nezha-agent.service <<EOF
[Unit]
Description=Nezha Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/nezha-agent -s $DASHBOARD_HOST -p $AGENT_SECRET
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
echo "启动服务..."
systemctl daemon-reload
systemctl enable nezha-agent
systemctl start nezha-agent

echo ""
echo "=== 安装完成 ==="
echo "查看状态: systemctl status nezha-agent"
echo "查看日志: journalctl -u nezha-agent -f"

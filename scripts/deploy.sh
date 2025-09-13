#!/bin/bash

# FileCodeBox Go版本部署脚本

set -e

# 配置
APP_NAME="filecodebox"
SERVICE_PORT="12345"
DATA_DIR="./data"
LOG_DIR="./logs"

echo "=== FileCodeBox Go版本部署脚本 ==="
echo

# 创建必要目录
echo "1. 创建必要目录..."
mkdir -p $DATA_DIR $LOG_DIR
mkdir -p $DATA_DIR/share/data
mkdir -p themes/2025/assets

echo "2. 检查依赖..."
# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: Go环境未安装"
    exit 1
fi

echo "Go版本: $(go version)"

# 编译应用
echo "3. 编译应用..."
go build -o $APP_NAME -ldflags="-w -s" .
chmod +x $APP_NAME

echo "编译完成: $(ls -lh $APP_NAME)"

# 创建systemd服务文件 (Linux)
if command -v systemctl &> /dev/null; then
    echo "4. 创建systemd服务..."
    sudo tee /etc/systemd/system/$APP_NAME.service > /dev/null <<EOF
[Unit]
Description=FileCodeBox File Sharing Service
After=network.target

[Service]
Type=simple
User=nobody
Group=nobody
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/$APP_NAME
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

    echo "systemd服务文件已创建: /etc/systemd/system/$APP_NAME.service"
    echo "使用以下命令管理服务:"
    echo "  sudo systemctl start $APP_NAME"
    echo "  sudo systemctl enable $APP_NAME"
    echo "  sudo systemctl status $APP_NAME"
fi

# 创建启动脚本
echo "5. 创建启动脚本..."
cat > start.sh <<EOF
#!/bin/bash
cd "\$(dirname "\$0")"
nohup ./$APP_NAME > $LOG_DIR/app.log 2>&1 &
echo \$! > $APP_NAME.pid
echo "FileCodeBox已启动，PID: \$(cat $APP_NAME.pid)"
echo "访问地址: http://localhost:$SERVICE_PORT"
echo "日志文件: $LOG_DIR/app.log"
EOF

chmod +x start.sh

# 创建停止脚本
cat > stop.sh <<EOF
#!/bin/bash
cd "\$(dirname "\$0")"
if [ -f $APP_NAME.pid ]; then
    PID=\$(cat $APP_NAME.pid)
    kill \$PID
    rm -f $APP_NAME.pid
    echo "FileCodeBox已停止"
else
    echo "PID文件不存在，尝试通过进程名停止..."
    pkill $APP_NAME
fi
EOF

chmod +x stop.sh

# 创建重启脚本
cat > restart.sh <<EOF
#!/bin/bash
cd "\$(dirname "\$0")"
./stop.sh
sleep 2
./start.sh
EOF

chmod +x restart.sh

# 创建状态检查脚本
cat > status.sh <<EOF
#!/bin/bash
cd "\$(dirname "\$0")"

if [ -f $APP_NAME.pid ]; then
    PID=\$(cat $APP_NAME.pid)
    if ps -p \$PID > /dev/null; then
        echo "FileCodeBox正在运行，PID: \$PID"
        echo "端口状态:"
        lsof -i :$SERVICE_PORT || netstat -tlnp | grep :$SERVICE_PORT
        echo
        echo "内存使用:"
        ps -p \$PID -o pid,pcpu,pmem,rss,vsz,comm
        echo
        echo "最近日志:"
        tail -n 10 $LOG_DIR/app.log
    else
        echo "FileCodeBox进程不存在，PID文件已过期"
        rm -f $APP_NAME.pid
    fi
else
    echo "FileCodeBox未运行"
fi
EOF

chmod +x status.sh

# 创建日志清理脚本
cat > cleanup.sh <<EOF
#!/bin/bash
cd "\$(dirname "\$0")"

echo "清理日志文件..."
find $LOG_DIR -name "*.log" -mtime +7 -delete

echo "清理过期数据..."
./$APP_NAME admin clean 2>/dev/null || echo "使用管理API清理..."

echo "清理完成"
EOF

chmod +x cleanup.sh

echo "6. 配置文件检查..."
if [ ! -f "$DATA_DIR/config.json" ]; then
    echo "配置文件将在首次运行时自动生成"
else
    echo "配置文件已存在: $DATA_DIR/config.json"
fi

echo
echo "=== 部署完成 ==="
echo
echo "可用的管理脚本:"
echo "  ./start.sh    - 启动服务"
echo "  ./stop.sh     - 停止服务"
echo "  ./restart.sh  - 重启服务"
echo "  ./status.sh   - 查看状态"
echo "  ./cleanup.sh  - 清理日志和过期数据"
echo
echo "配置目录: $DATA_DIR"
echo "日志目录: $LOG_DIR"
echo "访问地址: http://localhost:$SERVICE_PORT"
echo
echo "现在可以运行 './start.sh' 启动服务"

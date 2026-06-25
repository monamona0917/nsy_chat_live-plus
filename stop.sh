# 查找名为replive 和 bootstrap_rep 的两个进程然后kill掉进程
#!/bin/bash

# 查找名为replive 的进程
replive_pid=$(pgrep -f "replive")

# 查找名为bootstrap_rep 的进程
bootstrap_rep_pid=$(pgrep -f "bootstrap_rep")

# 检查是否找到进程
if [ -n "$replive_pid" ]; then
    echo "找到进程 replive，PID 为 $replive_pid"
    # 终止进程
    kill -9 "$replive_pid"
    echo "进程 replive 已终止"
else
    echo "未找到进程 replive"
fi

# 检查是否找到进程
if [ -n "$bootstrap_rep_pid" ]; then
    echo "找到进程 bootstrap_rep，PID 为 $bootstrap_rep_pid"
    # 终止进程
    kill "$bootstrap_rep_pid"
    echo "进程 bootstrap_rep 已终止"
else
    echo "未找到进程 bootstrap_rep"
fi

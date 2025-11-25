# Nezha Agent Terminal Audit Configuration
# 此文件在 bash 启动时被 source，用于设置审计钩子

# 获取环境变量
AUDIT_API_URL="${NEZHA_AUDIT_API_URL:-}"
STREAM_ID="${NEZHA_STREAM_ID:-}"

# 如果审计未启用，直接返回
if [ -z "$AUDIT_API_URL" ] || [ -z "$STREAM_ID" ]; then
    return 0
fi

# 空操作函数（用于替换被拦截的命令）
__nezha_noop() {
    return 1
}

# 生成6位随机验证码
__nezha_generate_code() {
    echo $(( RANDOM % 900000 + 100000 ))
}

# 检查命令函数
__nezha_check_command() {
    local cmd="$1"
    local cwd="$2"

    # 调用本地 API 检查命令
    local response=$(curl -s -X POST "$AUDIT_API_URL/check-command" \
        -H "Content-Type: application/json" \
        -d "{\"stream_id\":\"$STREAM_ID\",\"command\":\"$cmd\",\"working_dir\":\"$cwd\"}" \
        --max-time 2 2>/dev/null)

    # 如果 API 调用失败，允许执行
    if [ $? -ne 0 ] || [ -z "$response" ]; then
        return 0
    fi

    # 解析 JSON 响应
    local success=$(echo "$response" | grep -o '"success":[^,}]*' | cut -d':' -f2 | tr -d ' ')
    if [ "$success" != "true" ]; then
        return 0
    fi

    # 从data中提取blocked和reason
    local data=$(echo "$response" | grep -o '"data":{[^}]*}' | sed 's/"data"://g')
    local blocked=$(echo "$data" | grep -o '"blocked":[^,}]*' | cut -d':' -f2 | tr -d ' ')
    local reason=$(echo "$data" | grep -o '"reason":"[^"]*"' | cut -d'"' -f4)
    local action=$(echo "$data" | grep -o '"action":"[^"]*"' | cut -d'"' -f4)

    # 如果命令被拦截
    if [ "$blocked" = "true" ]; then
        echo -e "\033[31m✗ 命令被拦截: $reason\033[0m" >&2
        return 1
    fi

    # 如果是警告模式，需要验证码
    if [ "$action" = "warn" ] && [ -n "$reason" ]; then
        local code=$(__nezha_generate_code)
        echo -e "\033[33m⚠ 警告: $reason\033[0m" >&2
        echo -e "\033[33m如需继续执行，请输入验证码: \033[1m$code\033[0m" >&2
        read -p "验证码: " user_code
        if [ "$user_code" != "$code" ]; then
            echo -e "\033[31m✗ 验证码错误，命令已取消\033[0m" >&2
            return 1
        fi
        echo -e "\033[32m✓ 验证成功，继续执行\033[0m" >&2
    fi

    return 0
}

# 记录命令函数
__nezha_record_command() {
    local cmd="$1"
    local cwd="$2"
    local exit_code="$3"

    # 异步记录命令（后台执行）
    (curl -s -X POST "$AUDIT_API_URL/record-command" \
        -H "Content-Type: application/json" \
        -d "{\"stream_id\":\"$STREAM_ID\",\"command\":\"$cmd\",\"working_dir\":\"$cwd\",\"exit_code\":$exit_code}" \
        --max-time 2 >/dev/null 2>&1 &)
}

# 设置 PROMPT_COMMAND 来记录命令
__nezha_prompt_command() {
    local last_exit=$?
    # 获取最后执行的命令
    local last_cmd=$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')

    # 如果命令不为空且与上次不同
    if [ -n "$last_cmd" ] && [ "$last_cmd" != "$__NEZHA_PREV_CMD" ]; then
        __nezha_record_command "$last_cmd" "$PWD" "$last_exit"
        __NEZHA_PREV_CMD="$last_cmd"
    fi
}

# 设置 preexec 函数来检查命令
__nezha_preexec() {
    local cmd="$BASH_COMMAND"

    # 跳过 PROMPT_COMMAND 和内部命令
    if [[ "$cmd" == "__nezha_prompt_command" ]] || [[ "$cmd" == "__nezha_"* ]]; then
        return 0
    fi

    # 检查命令是否被拦截
    if ! __nezha_check_command "$cmd" "$PWD"; then
        # 在 extdebug 模式下，返回 1 会跳过命令执行
        return 1
    fi
}

# 启用命令历史
set -o history

# 加载用户的bashrc以获取aliases和其他配置
if [ -f "$HOME/.bashrc" ]; then
    source "$HOME/.bashrc"
fi

# 启用 extdebug 模式 - 这样 DEBUG trap 返回非零值会跳过命令
shopt -s extdebug

# 设置 PROMPT_COMMAND
if [ -z "$PROMPT_COMMAND" ]; then
    PROMPT_COMMAND="__nezha_prompt_command"
else
    PROMPT_COMMAND="__nezha_prompt_command; $PROMPT_COMMAND"
fi

# 设置 DEBUG trap 来拦截命令
trap '__nezha_preexec' DEBUG

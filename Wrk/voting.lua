local ticket

-- 初始化函数，读取票据值
function init()
    local file = io.open("current_ticket.txt", "r") -- 打开文件
    if file then
        ticket = file:read("*all"):gsub("%s+", "") -- 读取文件内容并去除可能的空白字符
        file:close() -- 关闭文件
    else
        ticket = "default_ticket" -- 如果文件不存在，使用默认票据值
    end
end

-- 请求函数
request = function()
    -- 包含多个用户名的投票请求
    -- local body = '{"query":"mutation { vote(name: [\\"Alice\\", \\"Bob\\"], ticket: \\"' .. ticket .. '\\") }"}'
    local body = '{"query":"mutation { vote(name: [\\"Alice\\"], ticket: \\"' .. ticket .. '\\") }"}'
    local headers = {
        ["Content-Type"] = "application/json"
    }
    return wrk.format("POST", nil, headers, body)
end

-- 确保在wrk运行之前调用init函数初始化票据
init()


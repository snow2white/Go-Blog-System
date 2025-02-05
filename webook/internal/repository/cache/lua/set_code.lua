-- local key = KEYS[1]
-- local cntKey = key..":cnt"
-- -- 你准备的存储的验证码
-- local val = ARGV[1]

-- local ttl = tonumber(redis.call("ttl", key))
-- if ttl == -1 then
--     --    key 存在，但是没有过期时间
--     return -2
-- elseif ttl == -2 or ttl < 540 then
--     --    可以发验证码
--     redis.call("set", key, val)
--     -- 600 秒
--     redis.call("expire", key, 600)
--     redis.call("set", cntKey, 3)
--     redis.call("expire", cntKey, 600)
--     return 0
-- else
--     -- 发送太频繁
--     return -1
-- end
local key = KEYS[1]
local cntKey = key..":cnt"
-- 你准备的存储的验证码
local val = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))
local cnt = tonumber(redis.call("get", cntKey))

if ttl == -1 then
    -- key 存在，但是没有过期时间
    return -2
elseif ttl == -2 or ttl < 599 then
    if cnt == nil or cnt > 0 then
        -- 可以发验证码
        redis.call("set", key, val)
        -- 600 秒
        redis.call("expire", key, 600)
        if cnt == nil then
            -- 初始化计数器为 3
            redis.call("set", cntKey, 3)
        else
            -- 递减计数器
            redis.call("decr", cntKey)
        end
        redis.call("expire", cntKey, 600)
        return 0
    else
        -- 计数器为 0，发送次数已达到上限
        return -1
    end
else
    -- 发送太频繁
    return -1
end
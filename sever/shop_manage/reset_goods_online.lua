--redis.call("scan","0","match","goods_info_*")

local goodis_key = redis.call("keys","goods_info_*")
if #goodis_key > 0 then
    redis.call("del",unpack(goodis_key))
end

-- KEYS = {"goods_info_23", "goods_info_24"}
-- ARGV = {
--     "name","lijun","id","123","desc","test is good","quantity","4",
--     "name","lijun1","id","444","desc","test is good1","quantity","5"
-- }

local start,res
local keys_len = table.getn(KEYS)
local argv_len = table.getn(ARGV)
local keys_argv_step = argv_len / keys_len
redis.debug(keys_len)

for i = 1,keys_len do
    start = (i - 1) * keys_argv_step + 1
    res = redis.call("hset", KEYS[i],unpack(ARGV, start, start + keys_argv_step - 1))
    redis.debug(res)
    if res == false then
        return false
    end
end

return true


--redis.debug(goodis_key)
--local name = redis.call("get","name")
--redis.debug(name)
--redis.debug(goods_keys)
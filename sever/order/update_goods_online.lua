--调试命令:redis-cli -a 密码 --ldb --eval lua文件路径 key1 key2 , arg1 arg2
redis.debug(KEYS[1])
local quantity = redis.call("hget", KEYS[1], "quantity")
redis.debug(quantity)
if quantity == false then
    return 0
end
redis.debug(ARGV[1])
if (tonumber(quantity) < tonumber(ARGV[1])) then
    return -1
end

local cur_quantity = quantity - ARGV[1]
redis.debug(cur_quantity)
redis.call("hset", KEYS[1], "quantity", cur_quantity)
return cur_quantity




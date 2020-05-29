DELIMITER $$

create procedure reponse_order_proc(IN goods_id int, in buy_num int)
BEGIN
	DECLARE cur_qiantity int;
    start transaction;
	set cur_qiantity = (select quantity from goods_info where id=goods_id for update);
    if (cur_qiantity is null) then
		set cur_qiantity=-2;
    else 
		if (cur_qiantity >= buy_num) then
			set cur_qiantity = cur_qiantity-buy_num;
			update goods_info set quantity=cur_qiantity where id=goods_id; 
		else
			set cur_qiantity=-1;
		end if;
    end if;
    commit;
    select cur_qiantity;
end $$

DELIMITER ;

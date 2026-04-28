-- ============ 第1层：矿石 ============
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level) VALUES
(1001, '铁陨石', 'ore', '最常见的小行星矿石，含有大量工业铁', 0.1, 10, true, 1),
(1002, '铜辉矿', 'ore', '含铜量较高的矿石，偶有银的伴生', 0.1, 15, true, 1),
(1003, '钛铁矿', 'ore', '钛和铁的共生矿物，轻质高强', 0.1, 25, true, 1),
(1004, '铬尖晶石', 'ore', '铬和镁的复合矿石', 0.1, 30, true, 1),
(1005, '钨锰矿', 'ore', '极高熔点的钨锰矿石，零安专属', 0.1, 80, true, 1),
(1006, '铌钽铁矿', 'ore', '含稀有金属铌和钽，超导材料来源', 0.1, 120, true, 1),
(1007, '钼铅矿', 'ore', '含钼和铅，高温结构和辐射屏蔽用', 0.1, 100, true, 1),
(1008, '铂族砂矿', 'ore', '含铂铱锇等贵金属的砂矿', 0.1, 250, true, 1),
(1009, '铪锆英石', 'ore', '极稀有矿石，含铪和锆', 0.1, 400, true, 1),
(1010, '锕系重矿', 'ore', '含铀钍等放射性元素的矿石', 0.1, 500, true, 1);

-- 气体
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level) VALUES
(1101, '氢雾', 'gas', '最常见的星云气体', 0.05, 5, true, 1),
(1102, '氦辉云', 'gas', '含氦-3的星云气体，聚变燃料来源', 0.05, 50, true, 1),
(1103, '碳氢星云', 'gas', '含甲烷和乙烯链的气体', 0.05, 20, true, 1);

-- 冰矿
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level) VALUES
(1201, '水冰', 'ice', '纯水冰，生命维持基础', 0.2, 8, true, 1),
(1202, '重水冰', 'ice', '含重水的冰块，聚变燃料', 0.2, 60, true, 1),
(1203, '氨冰', 'ice', '液氨冰块，冷却剂原料', 0.2, 15, true, 1);

-- ============ 第2层：精炼产出金属 ============
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level) VALUES
(2001, '工业纯铁', 'mineral', '最基础的结构材料', 0.01, 20, true, 2),
(2002, '电解铜', 'mineral', '高纯度铜，导线和电路基材', 0.01, 30, true, 2),
(2003, '海绵钛', 'mineral', '轻质高强结构材料', 0.01, 50, true, 2),
(2004, '金属铬', 'mineral', '耐腐蚀镀层材料', 0.01, 60, true, 2),
(2005, '钨粉', 'mineral', '极高熔点，发动机喷口材料', 0.01, 160, true, 2),
(2006, '铌条', 'mineral', '超导材料基础', 0.01, 240, true, 2),
(2007, '钼板', 'mineral', '高温结构材料', 0.01, 200, true, 2),
(2008, '铂', 'mineral', '贵金属催化剂', 0.01, 500, true, 2),
(2009, '电解锰', 'mineral', '合金添加剂', 0.01, 40, true, 2),
(2010, '镁锭', 'mineral', '轻合金基材', 0.01, 35, true, 2),
(2011, '燃料级氢', 'mineral', '常规推进燃料', 0.02, 10, true, 2),
(2012, '聚变级氦-3', 'mineral', '高效聚变燃料，极高价值', 0.02, 100, true, 2),
(2013, '超纯水', 'mineral', '冷却/化工用超纯水', 0.02, 15, true, 2);

-- ============ 第3层：合金/复合材料 ============
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level) VALUES
(3001, '船用结构钢', 'alloy', '舰船最基础的结构材料，用量极大', 0.05, 80, true, 3),
(3002, '航天钛合金', 'alloy', '轻质高强，护卫舰到巡洋舰的主结构', 0.05, 150, true, 3),
(3003, '重型钨钼板', 'alloy', '极端耐热，发动机和武器管壁', 0.05, 400, true, 3),
(3004, '铌基超导线材', 'alloy', '所有电磁系统的基础导线', 0.02, 500, true, 3),
(3005, '辐射屏蔽铅钨层', 'alloy', '反应堆周围的辐射屏蔽', 0.08, 300, true, 3),
(3006, '标准燃料棒', 'consumable', '裂变反应堆燃料', 0.1, 200, true, 3),
(3007, '聚变燃料球', 'consumable', '聚变堆弹丸', 0.05, 350, true, 3);

-- ============ 基础装备模块 ============
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level) VALUES
(5001, '小型采矿激光 I', 'module', '基础采矿装备', 5.0, 5000, false, 1),
(5002, '小型脉冲激光 I', 'module', '短程能量武器', 5.0, 8000, false, 1),
(5003, '小型护盾增强器 I', 'module', '加速护盾回充', 5.0, 6000, false, 1),
(5004, '小型装甲维修器 I', 'module', '修复装甲损伤', 5.0, 6000, false, 1),
(5005, '1MN加力推进器', 'module', '增加亚光速速度', 5.0, 4000, false, 1),
(5006, '磁轨炮 I', 'module', '中程动能武器', 5.0, 10000, false, 1),
(5007, '导弹发射器 I', 'module', '远程爆炸武器', 5.0, 12000, false, 1),
(5008, '跃迁干扰器 I', 'module', '阻止目标跃迁逃跑', 5.0, 15000, false, 1);

-- ============ 精炼配方 ============
INSERT INTO refine_recipes (input_item_id, input_quantity, output_item_id, output_quantity) VALUES
(1001, 100, 2001, 50),  -- 铁陨石 -> 工业纯铁
(1002, 100, 2002, 40),  -- 铜辉矿 -> 电解铜
(1003, 100, 2003, 30),  -- 钛铁矿 -> 海绵钛
(1003, 100, 2001, 20),  -- 钛铁矿 -> 工业纯铁(副产)
(1004, 100, 2004, 30),  -- 铬尖晶石 -> 金属铬
(1004, 100, 2010, 25),  -- 铬尖晶石 -> 镁锭(副产)
(1005, 100, 2005, 20),  -- 钨锰矿 -> 钨粉
(1005, 100, 2009, 15),  -- 钨锰矿 -> 电解锰(副产)
(1006, 100, 2006, 15),  -- 铌钽铁矿 -> 铌条
(1007, 100, 2007, 20),  -- 钼铅矿 -> 钼板
(1008, 100, 2008, 10),  -- 铂族砂矿 -> 铂
(1101, 100, 2011, 80),  -- 氢雾 -> 燃料级氢
(1102, 100, 2012, 30),  -- 氦辉云 -> 聚变级氦-3
(1201, 100, 2013, 60);  -- 水冰 -> 超纯水

-- ============ 制造配方：第3层合金 ============
INSERT INTO manufacture_recipes (id, output_item_id, output_quantity, build_time_sec, tech_level) VALUES
(1, 3001, 10, 120, 3),   -- 船用结构钢
(2, 3002, 5, 180, 3),    -- 航天钛合金
(3, 3003, 2, 300, 3),    -- 重型钨钼板
(4, 3004, 5, 240, 3),    -- 铌基超导线材
(5, 3005, 3, 200, 3),    -- 辐射屏蔽铅钨层
(6, 3006, 10, 60, 3),    -- 标准燃料棒
(7, 3007, 5, 90, 3);     -- 聚变燃料球

-- 制造材料：船用结构钢 = 工业纯铁50 + 电解锰10 + 金属铬5
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(1, 2001, 50), (1, 2009, 10), (1, 2004, 5);

-- 航天钛合金 = 海绵钛30 + 镁锭10
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(2, 2003, 30), (2, 2010, 10);

-- 重型钨钼板 = 钨粉20 + 钼板10
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(3, 2005, 20), (3, 2007, 10);

-- 铌基超导线材 = 铌条15 + 海绵钛5
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(4, 2006, 15), (4, 2003, 5);

-- 辐射屏蔽铅钨层 = 钨粉10 + 工业纯铁30
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(5, 2005, 10), (5, 2001, 30);

-- 标准燃料棒 = 燃料级氢50 + 超纯水20
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(6, 2011, 50), (6, 2013, 20);

-- 聚变燃料球 = 聚变级氦-3 20 + 超纯水10
INSERT INTO manufacture_materials (recipe_id, item_def_id, quantity) VALUES
(7, 2012, 20), (7, 2013, 10);

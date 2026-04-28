-- ============ 完善技能树：从27个扩展到75个 ============

-- ======== 驾驶技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(6, '无畏舰驾驶', 'piloting', '驾驶T6无畏舰', 10, 'perception', 'willpower', 5, 4),
(7, '航母驾驶', 'piloting', '驾驶T6航母', 10, 'perception', 'willpower', 5, 4),
(8, '泰坦驾驶', 'piloting', '驾驶T7泰坦级旗舰', 14, 'perception', 'willpower', 6, 4),
(9, '种族舰船适配', 'piloting', '每级减少5%操控非本族舰船的效率惩罚', 3, 'perception', 'charisma', 3, 3),
(60, '后勤航驾驶', 'piloting', '驾驶T6后勤航', 10, 'perception', 'willpower', 7, 3),
(61, '超级航母驾驶', 'piloting', '驾驶T6超级航母', 12, 'perception', 'willpower', 7, 4),
(62, '世界舰操作', 'piloting', '操作T8世界舰级设施', 16, 'perception', 'willpower', 8, 5),
(63, '飞船敏捷', 'piloting', '每级减少5%对齐时间', 2, 'perception', 'willpower', NULL, 0);

-- ======== 武器技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(16, '大型能量武器', 'gunnery', '操作大型激光/等离子武器', 5, 'perception', 'memory', 13, 4),
(17, '大型混合武器', 'gunnery', '操作大型磁轨/离子武器', 5, 'perception', 'memory', 14, 4),
(18, '大型导弹', 'gunnery', '操作大型导弹/鱼雷发射器', 5, 'perception', 'memory', 12, 4),
(19, '小型电磁武器', 'gunnery', '操作小型电磁脉冲武器', 1, 'perception', 'memory', NULL, 0),
(64, '中型电磁武器', 'gunnery', '操作中型电磁武器', 3, 'perception', 'memory', 19, 3),
(65, '大型电磁武器', 'gunnery', '操作大型电磁武器', 5, 'perception', 'memory', 64, 4),
(66, '无人机操控', 'gunnery', '每级+1可同时控制的无人机数', 1, 'memory', 'perception', NULL, 0),
(67, '无人机精通', 'gunnery', '每级+10%无人机伤害', 3, 'memory', 'perception', 66, 4),
(68, '武器超载', 'gunnery', '每级减少10%超载过热伤害', 4, 'perception', 'willpower', 15, 3),
(69, '弹道学', 'gunnery', '每级+5%所有武器最佳射程', 3, 'perception', 'intelligence', 15, 2),
(70, '速射', 'gunnery', '每级减少3%武器循环时间', 3, 'perception', 'memory', 15, 2),
(71, '奇异武器操作', 'gunnery', '操作第7层+奇异/概率武器', 8, 'perception', 'intelligence', 16, 5),
(72, '终极武器操作', 'gunnery', '操作泰坦末日武器', 12, 'perception', 'willpower', 71, 4);

-- ======== 防御技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(24, '护盾抗性强化', 'defense', '每级+3%护盾全抗性', 3, 'intelligence', 'memory', 20, 3),
(25, '装甲抗性强化', 'defense', '每级+3%装甲全抗性', 3, 'intelligence', 'memory', 22, 3),
(26, '结构工程学', 'defense', '每级+5%结构HP', 2, 'intelligence', 'memory', NULL, 0),
(27, '维修系统', 'defense', '每级+5%装甲维修量', 2, 'intelligence', 'memory', 22, 2),
(28, '遥修技术', 'defense', '每级+5%遥修距离和效率', 3, 'intelligence', 'memory', 27, 3),
(29, '电容管理', 'defense', '每级+5%电容总量', 2, 'intelligence', 'memory', NULL, 0),
(73, '电容回充', 'defense', '每级+5%电容回充速度', 2, 'intelligence', 'memory', 29, 2),
(74, '概率偏转操作', 'defense', '操作概率偏转护盾(T7+)', 8, 'intelligence', 'willpower', 24, 5),
(75, '因果屏障操作', 'defense', '操作因果断裂屏障(T9+)', 12, 'intelligence', 'willpower', 74, 4);

-- ======== 工业技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(34, '高级精炼', 'industry', '每级+3%稀有矿精炼效率', 4, 'memory', 'intelligence', 31, 4),
(35, '高级制造', 'industry', '每级减少3%大型舰船制造时间', 4, 'memory', 'intelligence', 32, 4),
(36, '行星管理', 'industry', '每级+1可管理的行星数', 3, 'memory', 'intelligence', NULL, 0),
(37, '气体采集', 'industry', '操作气体采集器', 2, 'memory', 'intelligence', 30, 2),
(38, '冰矿开采', 'industry', '操作冰矿采集器', 2, 'memory', 'intelligence', 30, 2),
(39, '现象级采集操作', 'industry', '操作第4层现象级采集设备', 8, 'memory', 'intelligence', 34, 4),
(76, '拓扑编织操作', 'industry', '操作第6层拓扑编织设备', 10, 'memory', 'intelligence', 39, 4),
(77, '超材料锻造操作', 'industry', '操作第7层超材料锻造', 12, 'memory', 'intelligence', 76, 4),
(78, '打捞', 'industry', '从残骸中回收材料和组件', 1, 'memory', 'perception', NULL, 0),
(79, '逆向工程', 'industry', '从残骸中提取科技蓝图', 4, 'memory', 'intelligence', 78, 3),
(80, '发明', 'industry', '将T1蓝图发明为T2版本', 5, 'memory', 'intelligence', 33, 4);

-- ======== 导航技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(43, '跳跃引擎操作', 'navigation', '操作跳跃引擎(主力舰级)', 6, 'intelligence', 'perception', 42, 4),
(44, '跳跃引擎校准', 'navigation', '每级+10%跳跃距离', 5, 'intelligence', 'perception', 43, 3),
(45, '微跃引擎', 'navigation', '每级减少5%微跃充能时间', 3, 'intelligence', 'perception', 40, 3),
(46, '深空感知', 'navigation', '每级+10%探针扫描强度', 2, 'intelligence', 'perception', NULL, 0),
(47, '星图学', 'navigation', '每级+5%扫描精度和速度', 3, 'intelligence', 'perception', 46, 2),
(81, '虫洞导航', 'navigation', '每级减少10%虫洞穿越风险', 4, 'intelligence', 'perception', 47, 3),
(82, '维度滑移', 'navigation', '操作维度滑移引擎(T9+)', 10, 'intelligence', 'perception', 44, 5);

-- ======== 领导力（全新分类） ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(83, '小队指挥', 'leadership', '每级+2%小队(10人)属性加成', 2, 'charisma', 'willpower', NULL, 0),
(84, '舰队指挥', 'leadership', '每级+2%舰队(100人)属性加成', 4, 'charisma', 'willpower', 83, 4),
(85, '大舰队指挥', 'leadership', '每级+2%大舰队(1000人)属性加成', 6, 'charisma', 'willpower', 84, 4),
(86, '装甲指挥', 'leadership', '舰队装甲HP+每级3%', 3, 'charisma', 'willpower', 83, 2),
(87, '护盾指挥', 'leadership', '舰队护盾HP+每级3%', 3, 'charisma', 'willpower', 83, 2),
(88, '信息指挥', 'leadership', '舰队传感器强度+每级3%', 3, 'charisma', 'willpower', 83, 2),
(89, '意识广播指挥', 'leadership', '通过意识网络指挥(T9+)', 10, 'charisma', 'intelligence', 85, 4),
(90, '战争动员', 'leadership', '每级减少5%舰队成员跃迁对齐时间', 4, 'charisma', 'willpower', 84, 3);

-- ======== 贸易技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(55, '订单管理', 'trade', '每级+5个同时挂单数', 2, 'charisma', 'memory', 54, 2),
(56, '合同谈判', 'trade', '每级减少5%合同手续费', 3, 'charisma', 'memory', 54, 3),
(57, '走私学', 'trade', '每级减少10%被巡逻检查概率', 3, 'charisma', 'willpower', 54, 2),
(58, '区域贸易', 'trade', '每级+1可远程挂单的星系跳数', 4, 'charisma', 'memory', 55, 3),
(59, '期货交易', 'trade', '每级减少5%期货保证金要求', 5, 'charisma', 'intelligence', 58, 4);

-- ======== 科研技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(91, '实验分析', 'research', '每级+5%从失败研究中获取经验', 2, 'intelligence', 'memory', 53, 2),
(92, '高等理论', 'research', '解锁第4-5层科技研发', 6, 'intelligence', 'memory', 53, 4),
(93, '拓扑物理学', 'research', '解锁第6层科技研发', 8, 'intelligence', 'memory', 92, 4),
(94, '时空几何学', 'research', '解锁第7层科技研发', 10, 'intelligence', 'memory', 93, 4),
(95, '因果律数学', 'research', '解锁第8-9层科技研发', 14, 'intelligence', 'memory', 94, 4);

-- ======== 特殊技能补全 ========
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr, prereq_skill_id, prereq_level) VALUES
(96, '意识碎片整合', 'special', '每级+5%与舰船AI的协调度', 4, 'intelligence', 'willpower', 51, 3),
(97, '种族亲和', 'special', '每级减少10%学习其他种族科技的时间', 3, 'charisma', 'perception', NULL, 0),
(98, '高级自动化', 'special', '解锁复杂if-then自动化规则', 5, 'intelligence', 'memory', 50, 3),
(99, '离线自动化', 'special', '解锁离线后角色继续执行任务', 6, 'intelligence', 'memory', 98, 4);

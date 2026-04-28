#!/usr/bin/env python3
"""
调用 DeepSeek V4 Pro API 生成星陨战歌全部252艘舰船的SQL数据
"""
import json
import os
import sys
import time
import requests

API_KEY = os.environ.get("DEEPSEEK_API_KEY", "")
API_URL = "https://api.deepseek.com/chat/completions"
MODEL = "deepseek-v4-pro"
OUTPUT_FILE = os.path.join(os.path.dirname(__file__), "..", "migrations", "013_all_ships.up.sql")

RACES = [
    (1, "白羊座·冲锋者", "火象", "高伤害、速度快、装甲薄，热能武器加成"),
    (2, "金牛座·铸造者", "土象", "重装甲、慢速、大货仓，装甲和采矿加成"),
    (3, "双子座·幻影者", "风象", "最快速度、电子战强、薄皮，信号半径小"),
    (4, "巨蟹座·守护者", "水象", "护盾最强、维修加成、速度中等"),
    (5, "狮子座·统御者", "火象", "旗舰指挥加成、均衡偏攻、中等速度"),
    (6, "处女座·精工者", "土象", "电子战精确、传感器强、制造加成"),
    (7, "天秤座·裁量者", "风象", "后勤支援、均衡设计、货仓大"),
    (8, "天蝎座·蚀刻者", "水象", "暗杀专精、持续伤害、能量吸取"),
    (9, "射手座·游猎者", "火象", "远程狙击、高机动、投射武器"),
    (10, "摩羯座·筑垒者", "土象", "堡垒防御、结构血量极高、最慢"),
    (11, "水瓶座·革新者", "风象", "实验武器、非常规加成、CPU高"),
    (12, "双鱼座·共生者", "水象", "生物科技、自我修复、适应性强"),
]

SHIP_TEMPLATE = """
每个种族需要生成以下21艘舰船的SQL INSERT:
T1护卫舰(4艘): 突击型/电子型/后勤型/工业型
T2驱逐舰(3艘): 火力型/防御型/拦截型
T3巡洋舰(5艘): 主战型/电子战型/后勤型/采矿型/隐形型
T4战列巡洋舰(3艘): 攻击型/指挥型/重火力型
T5战列舰(3艘): 主力型/狙击型/近战型
T6无畏舰(1艘) + 航母(1艘)
T7泰坦(1艘)

数值基准范围(请严格遵守):
T1护卫: 护盾400-900, 装甲300-900, 结构300-500, 速度250-500, 对齐2-4, 信号30-55, 高槽2-3, 中槽2-3, 低槽2-3, PG40-70, CPU100-170
T2驱逐: 护盾1200-1800, 装甲800-1500, 结构600-1000, 速度250-400, 对齐3-5, 信号60-100, 高槽4-6, 中槽2-4, 低槽2-3, PG80-100, CPU160-200
T3巡洋: 护盾2500-5000, 装甲2000-5500, 结构1500-3000, 速度180-300, 对齐4-7, 信号100-200, 高槽3-5, 中槽3-5, 低槽3-5, PG130-200, CPU220-320
T4战巡: 护盾5000-8000, 装甲4000-8000, 结构3000-5000, 速度150-250, 对齐5-8, 信号180-280, 高槽5-7, 中槽3-5, 低槽3-5, PG200-300, CPU280-380
T5战列: 护盾8000-14000, 装甲6000-12000, 结构4000-8000, 速度100-180, 对齐7-10, 信号250-400, 高槽6-8, 中槽4-6, 低槽4-6, PG350-600, CPU350-500
T6主力舰: 护盾20000-40000, 装甲15000-35000, 结构10000-20000, 速度60-120, 对齐10-15, 信号500-800, 高槽5-8, 中槽4-6, 低槽4-6, PG800-1500, CPU600-1000
T7泰坦: 护盾80000-120000, 装甲60000-100000, 结构40000-70000, 速度30-60, 对齐15-25, 信号1000-2000, 高槽8-10, 中槽6-8, 低槽6-8, PG2000-4000, CPU1500-2500

护盾回充=护盾/50, 电容=高槽*100+200, 电容回充=电容/40, 跃迁速度: T1=4-6, T2=3.5-4.5, T3=2.5-3.5, T4=2-3, T5=1.5-2.5, T6=1-1.5, T7=0.5-1
货仓: 突击100-200, 工业400-800, 采矿300-500, 其他200-500(按级别递增)

抗性(每种族不同，同族趋势一致):
火象(白羊/狮子/射手): 护盾热抗高(0.40-0.50)爆炸低(0.05-0.15), 装甲热抗高(0.40-0.50)电磁低(0.10-0.15)
土象(金牛/处女/摩羯): 护盾均衡(0.20), 装甲动能高(0.45-0.55)电磁低(0.10-0.20)
风象(双子/天秤/水瓶): 护盾电磁高(0.40-0.50)爆炸低(0.10), 装甲电磁高(0.35-0.45)爆炸低(0.10-0.15)
水象(巨蟹/天蝎/双鱼): 护盾爆炸高(0.35-0.45)动能低(0.10-0.15), 装甲均衡偏高(0.30-0.40)
"""

def generate_race_ships(race_id, race_name, element, race_desc, ship_names):
    prompt = f"""你是一个游戏数值设计师。请为星陨战歌游戏的 {race_name} 生成21艘舰船的SQL INSERT语句。

种族特点: {race_desc}
元素: {element}

已有的舰船名称(按顺序对应21种型号):
{json.dumps(ship_names, ensure_ascii=False)}

{SHIP_TEMPLATE}

请直接输出SQL INSERT语句，格式如下(不要解释，不要markdown标记，直接输出纯SQL):
INSERT INTO ship_defs (id, name, race_id, tier, ship_class, ship_role, shield_hp, armor_hp, structure_hp, shield_recharge, capacitor, cap_recharge, max_speed, warp_speed, warp_cap_cost, align_ticks, mass, signature, high_slots, mid_slots, low_slots, cargo_m3, drone_bay_m3, powergrid, cpu, jump_range, shield_res_kinetic, shield_res_thermal, shield_res_em, shield_res_explosive, armor_res_kinetic, armor_res_thermal, armor_res_em, armor_res_explosive, race_bonus) VALUES
(...每艘一行...);

id从 {race_id * 100 + 1} 开始递增。race_id = {race_id}。
race_bonus写中文，描述该船的种族技能加成(每级XX技能+X%某属性)。
"""
    
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {API_KEY}"
    }
    
    data = {
        "model": MODEL,
        "messages": [
            {"role": "system", "content": "你是一个专业的游戏数值设计师，精通EVE Online式太空游戏的舰船数值平衡。你只输出SQL代码，不做解释。"},
            {"role": "user", "content": prompt}
        ],
        "temperature": 0.3,
        "max_tokens": 8000,
        "stream": False
    }
    
    print(f"  正在生成 {race_name} 的21艘舰船...", flush=True)
    
    try:
        resp = requests.post(API_URL, headers=headers, json=data, timeout=120)
        resp.raise_for_status()
        result = resp.json()
        content = result["choices"][0]["message"]["content"]
        
        # 提取SQL部分
        sql = content.strip()
        if "```" in sql:
            parts = sql.split("```")
            for p in parts:
                if "INSERT" in p:
                    sql = p.replace("sql", "").strip()
                    break
        
        print(f"  ✓ {race_name} 完成", flush=True)
        return sql
    except Exception as e:
        print(f"  ✗ {race_name} 失败: {e}", flush=True)
        return f"-- {race_name} 生成失败: {e}\n"

# 12种族的舰船名称(来自设计文档附录I)
RACE_SHIP_NAMES = {
    1: ["锐矛","怒目","血誓","锻角","焚城","金毛","破阵","战嚎","血雾","军医","掘角","伏羊","裂阵","号角","烈焰","阿瑞斯","刺穿","羝撞","战神之怒","军团摇篮","金毛天驱"],
    2: ["铁蹄","炉眼","补给","犁刃","铁锤","牛盾","铁栅","弥诺","迷宫","丰收","掘金","暗矿","怒牛","锻炉","陨铁","赫菲斯托斯","投矛","铁角","大地之锤","母矿脉","盖亚之心"],
    3: ["影刺","迷瞳","双翼","拾遗","镜裂","幻壁","缠影","卡斯托尔","波鲁克斯","共息","暗面","虚影","分魂","双令","交叉","雅努斯","远影","合体","镜界崩塌","千面之巢","永恒双生"],
    4: ["钳击","月瞳","暖壳","拾贝","碎钳","甲壳","潮锁","塞勒涅","月蚀","母蟹","掘沙","潜渊","怒潮","潮令","碎壳","阿尔忒弥斯","银弦","坚壁","月神堡垒","潮汐摇篮","永夜守望"],
    5: ["利爪","威视","侍从","贡品","裂吼","鬃盾","围猎","涅墨亚","耀辉","王恩","王领","伏狮","征服","王座","日炎","赫利俄斯","神罚","狮王","日冕之怒","王庭","永恒王权"],
    6: ["锐针","析目","织补","采撷","裁断","秩序","缚丝","雅典娜","阿斯特瑞亚","缝合","筛选","无痕","审判","蓝图","解构","得墨忒尔","天衡","收割","完美方程","智慧神殿","星辰秩序"],
    7: ["衡击","契约","馈赠","称量","砝码","和约","关税","忒弥斯","诡辩","均分","采买","暗账","制裁","议长","仲裁","赫尔墨斯","远规","天秤","终审裁决","万商之港","黄金律法"],
    8: ["毒针","蚀眼","吸髓","拆骨","蚀骨","暗甲","绝路","塔纳托斯","倪克斯","汲魂","腐蚀","冥影","溶解","暗令","剧毒","哈迪斯","暗箭","吞噬","冥府裁判","万毒之巢","永夜深渊"],
    9: ["疾矢","鹰眼","驿马","采风","连射","奔逸","追风","喀戎","幻弦","草药","远征","猎伏","穿日","号令","万箭","阿波罗","天弓","践踏","猎神之怒","游牧营地","银河射线"],
    10: ["石锤","测绘","基石","凿岩","崩石","壁垒","落闸","阿特拉斯","封印","奠基","掘进","伏石","山崩","石令","攻城","克洛诺斯","石弩","铁壁","山岳之心","永固要塞","时间之墙"],
    11: ["电火","灵感","注能","试作","裂变","异构","超频","普罗米修斯","悖论","涌泉","萃取","相变","颠覆","先驱","聚变","乌拉诺斯","超限","突变","创世之火","范式革命","奇点突破"],
    12: ["噬咬","感潮","共生","滤食","漩涡","鳞甲","缠藻","普罗透斯","幻海","愈潮","珊瑚","深潜","海啸","洋流","巨口","波塞冬","三叉戟","利维坦","深海之心","孕海之母","万物归流"],
}

def main():
    if not API_KEY:
        print("错误: 需要设置 DEEPSEEK_API_KEY 环境变量")
        sys.exit(1)
    
    print("=== 星陨战歌 · 舰船数据生成器 ===")
    print(f"模型: {MODEL}")
    print(f"输出: {OUTPUT_FILE}")
    print(f"共需生成: 12种族 × 21艘 = 252艘舰船\n")
    
    all_sql = ["-- 星陨战歌全部252艘舰船数据 (由DeepSeek V4 Pro生成)\n"]
    all_sql.append("-- 生成时间: " + time.strftime("%Y-%m-%d %H:%M:%S") + "\n\n")
    all_sql.append("-- 先清理旧数据(保留已有的1-9号手动创建的船)\n")
    all_sql.append("DELETE FROM ship_defs WHERE id > 100;\n\n")
    
    for race_id, race_name, element, race_desc in RACES:
        ship_names = RACE_SHIP_NAMES[race_id]
        sql = generate_race_ships(race_id, race_name, element, race_desc, ship_names)
        all_sql.append(f"\n-- ============ {race_name} ============\n")
        all_sql.append(sql)
        all_sql.append("\n")
        time.sleep(1)  # 避免API限速
    
    with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
        f.write("\n".join(all_sql))
    
    print(f"\n=== 完成! SQL已保存到 {OUTPUT_FILE} ===")

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""补全缺失的舰船数据"""
import json, os, sys, time, requests

API_KEY = os.environ.get("DEEPSEEK_API_KEY", "")
API_URL = "https://api.deepseek.com/chat/completions"
OUTPUT = "/root/sws/migrations/015_fill_ships.up.sql"

MISSING = {
    1: {"name":"白羊座","need":"T4战巡(攻击/指挥/重火力)、T5战列(主力/狙击/近战)、T6无畏+航母、T7泰坦","names":"裂阵,号角,烈焰,阿瑞斯,刺穿,羝撞,战神之怒,军团摇篮,金毛天驱","start_id":113,"desc":"火象高伤害速度快，热能武器加成","resists":"shield:k0.22/t0.50/e0.20/x0.08 armor:k0.25/t0.45/e0.10/x0.20"},
    2: {"name":"金牛座","need":"T3巡洋(采矿/隐形)、T4战巡全部、T5战列全部、T6+T7","names":"掘金,暗矿,怒牛,锻炉,陨铁,赫菲斯托斯,投矛,铁角,大地之锤,母矿脉,盖亚之心","start_id":212,"desc":"土象重装甲大货仓，装甲和采矿加成","resists":"shield:k0.20/t0.25/e0.20/x0.15 armor:k0.50/t0.35/e0.25/x0.40"},
    3: {"name":"双子座","need":"T3巡洋(采矿/隐形)、T4战巡全部、T5战列全部、T6+T7","names":"暗面,虚影,分魂,双令,交叉,雅努斯,远影,合体,镜界崩塌,千面之巢,永恒双生","start_id":312,"desc":"风象最快电子战强薄皮，信号半径小","resists":"shield:k0.15/t0.20/e0.50/x0.10 armor:k0.20/t0.15/e0.40/x0.10"},
    4: {"name":"巨蟹座","need":"T5战列(主力/狙击/近战)、T6无畏+航母、T7泰坦","names":"阿尔忒弥斯,银弦,坚壁,月神堡垒,潮汐摇篮,永夜守望","start_id":416,"desc":"水象护盾最强维修加成，爆炸高抗","resists":"shield:k0.10/t0.25/e0.30/x0.40 armor:k0.30/t0.30/e0.25/x0.35"},
    5: {"name":"狮子座","need":"全部21艘","names":"利爪,威视,侍从,贡品,裂吼,鬃盾,围猎,涅墨亚,耀辉,王恩,王领,伏狮,征服,王座,日炎,赫利俄斯,神罚,狮王,日冕之怒,王庭,永恒王权","start_id":501,"desc":"火象旗舰指挥加成均衡偏攻","resists":"shield:k0.18/t0.48/e0.22/x0.08 armor:k0.28/t0.48/e0.12/x0.18"},
    7: {"name":"天秤座","need":"全部21艘","names":"衡击,契约,馈赠,称量,砝码,和约,关税,忒弥斯,诡辩,均分,采买,暗账,制裁,议长,仲裁,赫尔墨斯,远规,天秤,终审裁决,万商之港,黄金律法","start_id":701,"desc":"风象后勤支援均衡货仓大","resists":"shield:k0.18/t0.22/e0.45/x0.12 armor:k0.22/t0.18/e0.38/x0.12"},
    12: {"name":"双鱼座","need":"全部21艘","names":"噬咬,感潮,共生,滤食,漩涡,鳞甲,缠藻,普罗透斯,幻海,愈潮,珊瑚,深潜,海啸,洋流,巨口,波塞冬,三叉戟,利维坦,深海之心,孕海之母,万物归流","start_id":1201,"desc":"水象生物科技自我修复适应性强","resists":"shield:k0.12/t0.30/e0.25/x0.38 armor:k0.32/t0.35/e0.28/x0.32"},
}

# 还有部分缺的
PARTIAL = {
    8: {"name":"天蝎座","need":"T3巡洋(后勤/采矿/隐形)、T4战巡全部、T5战列全部、T6+T7","names":"汲魂,腐蚀,冥影,溶解,暗令,剧毒,哈迪斯,暗箭,吞噬,冥府裁判,万毒之巢,永夜深渊","start_id":810,"desc":"水象暗杀持续伤害能量吸取","resists":"shield:k0.12/t0.22/e0.28/x0.42 armor:k0.28/t0.28/e0.22/x0.38","count":14},
    9: {"name":"射手座","need":"T3巡洋(隐形)、T4战巡全部、T5战列全部、T6+T7","names":"猎伏,穿日,号令,万箭,阿波罗,天弓,践踏,猎神之怒,游牧营地,银河射线","start_id":912,"desc":"火象远程狙击高机动投射武器","resists":"shield:k0.20/t0.45/e0.18/x0.10 armor:k0.22/t0.42/e0.12/x0.22","count":11},
    10: {"name":"摩羯座","need":"T7泰坦、还缺的T6","names":"时间之墙,山岳之心,永固要塞","start_id":1019,"desc":"土象堡垒防御结构血量极高最慢","resists":"shield:k0.22/t0.28/e0.18/x0.18 armor:k0.52/t0.30/e0.18/x0.42","count":3},
    11: {"name":"水瓶座","need":"T3巡洋(后勤/采矿/隐形)、T4战巡全部、T5战列全部、T6+T7","names":"涌泉,萃取,相变,颠覆,先驱,聚变,乌拉诺斯,超限,突变,创世之火,范式革命,奇点突破","start_id":1110,"desc":"风象实验武器非常规加成CPU高","resists":"shield:k0.16/t0.18/e0.48/x0.10 armor:k0.18/t0.20/e0.42/x0.14","count":13},
}

ALL_MISSING = {**MISSING, **PARTIAL}

def generate(race_id, info):
    names = info["names"]
    prompt = f"""为星陨战歌的{info['name']}补全缺失舰船。需要: {info['need']}
舰船名称: {names}
种族特点: {info['desc']}
抗性: {info['resists']}

数值范围:
T3巡洋: 护盾2500-5000,装甲2000-5500,结构1500-3000,速度180-300,对齐4-7,信号100-200,高3-5,中3-5,低3-5,PG130-200,CPU220-320
T4战巡: 护盾5000-8000,装甲4000-8000,结构3000-5000,速度150-250,对齐5-8,信号180-280,高5-7,中3-5,低3-5,PG200-300,CPU280-380
T5战列: 护盾8000-14000,装甲6000-12000,结构4000-8000,速度100-180,对齐7-10,信号250-400,高6-8,中4-6,低4-6,PG350-600,CPU350-500
T6主力: 护盾20000-40000,装甲15000-35000,结构10000-20000,速度60-120,对齐10-15,信号500-800,高5-8,中4-6,低4-6,PG800-1500,CPU600-1000
T7泰坦: 护盾80000-120000,装甲60000-100000,结构40000-70000,速度30-60,对齐15-25,信号1000-2000,高8-10,中6-8,低6-8,PG2000-4000,CPU1500-2500

ID从{info['start_id']}开始。race_id={race_id}。直接输出SQL，每艘一个INSERT:
INSERT INTO ship_defs (id,name,race_id,tier,ship_class,ship_role,shield_hp,armor_hp,structure_hp,shield_recharge,capacitor,cap_recharge,max_speed,warp_speed,warp_cap_cost,align_ticks,mass,signature,high_slots,mid_slots,low_slots,cargo_m3,drone_bay_m3,powergrid,cpu,jump_range,shield_res_kinetic,shield_res_thermal,shield_res_em,shield_res_explosive,armor_res_kinetic,armor_res_thermal,armor_res_em,armor_res_explosive,race_bonus) VALUES (...);
"""
    headers = {"Content-Type":"application/json","Authorization":f"Bearer {API_KEY}"}
    data = {"model":"deepseek-v4-pro","messages":[
        {"role":"system","content":"你只输出SQL。每艘船一个完整的INSERT语句，以分号结尾。"},
        {"role":"user","content":prompt}
    ],"temperature":0.3,"max_tokens":6000,"stream":False}

    print(f"  生成 {info['name']}...", flush=True)
    try:
        r = requests.post(API_URL, headers=headers, json=data, timeout=180)
        r.raise_for_status()
        content = r.json()["choices"][0]["message"]["content"]
        sql = content.strip()
        if "```" in sql:
            for p in sql.split("```"):
                if "INSERT" in p: sql = p.replace("sql","").strip(); break
        print(f"  ✓ {info['name']} OK", flush=True)
        return sql
    except Exception as e:
        print(f"  ✗ {info['name']} 失败: {e}", flush=True)
        return f"-- {info['name']} 失败\n"

def main():
    if not API_KEY: print("需要DEEPSEEK_API_KEY"); sys.exit(1)
    print(f"=== 补全缺失舰船 ({len(ALL_MISSING)}个种族) ===\n")
    all_sql = ["-- 补全缺失舰船数据\n"]
    for rid in sorted(ALL_MISSING.keys()):
        info = ALL_MISSING[rid]
        sql = generate(rid, info)
        all_sql.append(f"\n-- {info['name']}\n{sql}\n")
        time.sleep(1)
    with open(OUTPUT, "w") as f: f.write("\n".join(all_sql))
    print(f"\n=== 完成! 保存到 {OUTPUT} ===")

if __name__=="__main__": main()

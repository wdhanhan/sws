package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/starfall-warsong/sws/pkg/config"
	"github.com/starfall-warsong/sws/pkg/database"
	"github.com/starfall-warsong/sws/pkg/logger"
)

type raceTheme struct {
	RaceID int
	Name   string
	Theme  string
	Bosses []string
	NPCIDs []int64 // 使用已有的NPC定义ID作为基础
}

var races = []raceTheme{
	{1, "白羊", "烈焰试炼场", []string{"焚天守卫", "战焰领主", "阿瑞斯之影"}, []int64{1, 2, 3}},
	{2, "金牛", "废弃矿井深渊", []string{"矿脉吞噬者", "铁心暴君", "盖亚碎片"}, []int64{1, 2, 4}},
	{3, "双子", "镜像迷宫", []string{"镜像分裂王", "虚实交错者", "永恒双面"}, []int64{1, 20, 21}},
	{4, "巨蟹", "潮汐防线废墟", []string{"潮汐堡垒核心", "月蚀巨蟹", "永夜守卫"}, []int64{1, 2, 3}},
	{5, "狮子", "皇家竞技场", []string{"篡位者", "堕落皇卫", "太阳裁决"}, []int64{2, 3, 4}},
	{6, "处女", "先驱者数据核心", []string{"完美计算者", "秩序执行体", "星辰方程"}, []int64{1, 2, 3}},
	{7, "天秤", "走私者巢穴", []string{"黑市之王", "暗法裁决者", "黄金天秤"}, []int64{2, 3, 11}},
	{8, "天蝎", "毒雾深渊", []string{"永夜蚀毒母", "深渊蝎王", "冥府审判"}, []int64{1, 20, 21}},
	{9, "射手", "星际猎场", []string{"星原追猎者", "虚空猎王", "银河之矢"}, []int64{10, 11, 12}},
	{10, "摩羯", "远古堡垒遗迹", []string{"时间石像", "永恒壁垒核心", "克洛诺斯残影"}, []int64{1, 4, 5}},
	{11, "水瓶", "实验事故区", []string{"奇点暴走体", "范式崩塌者", "创世余波"}, []int64{20, 21, 3}},
	{12, "双鱼", "深海生态巢穴", []string{"深渊利维坦", "共生母体", "万物潮汐"}, []int64{10, 11, 12}},
}

var diffNames = []string{"", "巡逻", "搜索", "突袭", "清剿", "歼灭", "攻坚", "远征", "深渊", "法则", "神格"}
var waveCounts = []int{0, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7}
var bossHPs = []int{0, 1500, 5000, 15000, 40000, 100000, 300000, 800000, 2000000, 5000000, 10000000}
var rewardCredits = []int64{0, 5000, 15000, 50000, 150000, 500000, 1500000, 5000000, 15000000, 50000000, 200000000}
var npcPerWave = []int{0, 1, 2, 3, 4, 4, 5, 6, 8, 10, 12}

var waveTexts = [][]string{
	{},
	{"一小队敌人正在巡逻，它们还没有发现你。", "更多的敌人从暗处涌出！", "★ 首领出现了！准备战斗！"},
	{"前方探测到中等规模的敌人编队。", "敌方增援抵达！小心侧翼！", "★ 区域守卫被惊动，精英单位出击！"},
	{"大量敌人信号——这里是它们的据点核心。", "防御系统启动，自动炮台开火！", "精英编队从四面八方包围！", "★ Boss战！据点核心守卫启动！"},
	{"你进入了敌方深层防线。", "第二道防线——重型单位出动。", "敌方调集了所有预备队！", "★ 高级指挥官亲自出战！"},
	{"这片区域弥漫着危险的气息...", "数十个敌方信号同时出现！", "精英小队发起了协调进攻！", "敌方的最强战力全部出动！", "★ 终极Boss——它已经等你很久了。"},
}

func main() {
	logger.Init("debug")
	cfg := config.Load()
	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	rng := rand.New(rand.NewSource(42))

	dungeonID := int64(1)
	totalWaves := 0

	for _, race := range races {
		for diff := 1; diff <= 10; diff++ {
			name := fmt.Sprintf("%s·%s [%s]", race.Name, race.Theme, diffNames[diff])
			desc := fmt.Sprintf("%s种族主题远征，难度%d(%s)，推荐T%d舰船", race.Name, diff, diffNames[diff], diff)
			wc := waveCounts[diff]

			_, err := db.ExecContext(ctx,
				`INSERT INTO dungeon_defs (id,name,description,race_theme,difficulty,wave_count,min_security,reward_credits,encounter_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (id) DO NOTHING`,
				dungeonID, name, desc, race.RaceID, diff, wc, -1.0, rewardCredits[diff], 0)
			if err != nil {
				log.Printf("dungeon %d: %v", dungeonID, err)
			}

			// Generate waves
			for w := 1; w <= wc; w++ {
				isBoss := w == wc
				npcID := race.NPCIDs[rng.Intn(len(race.NPCIDs))]
				count := npcPerWave[diff] + rng.Intn(3) - 1
				if count < 1 {
					count = 1
				}
				if isBoss {
					count = 1
				}

				bossName := ""
				bossHP := 0
				if isBoss {
					bossName = race.Bosses[rng.Intn(len(race.Bosses))]
					bossHP = bossHPs[diff]
				}

				waveTextIdx := w - 1
				texts := waveTexts[1] // default
				if diff <= 2 {
					texts = waveTexts[1]
				} else if diff <= 4 {
					texts = waveTexts[2]
				} else if diff <= 6 {
					texts = waveTexts[3]
				} else if diff <= 8 {
					texts = waveTexts[4]
				} else {
					texts = waveTexts[5]
				}
				wt := ""
				if waveTextIdx < len(texts) {
					wt = texts[waveTextIdx]
				} else {
					wt = texts[len(texts)-1]
				}

				db.ExecContext(ctx,
					`INSERT INTO dungeon_waves (dungeon_id,wave_number,npc_def_id,npc_count,is_boss,boss_name,boss_hp_override,wave_text)
					 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
					dungeonID, w, npcID, count, isBoss, bossName, bossHP, wt)
				totalWaves++
			}

			dungeonID++
		}
	}

	fmt.Printf("=== 远征副本生成完毕 ===\n")
	fmt.Printf("副本数: %d (12种族 x 10难度)\n", dungeonID-1)
	fmt.Printf("波次数: %d\n", totalWaves)
}

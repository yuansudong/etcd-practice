package center

type OnOff struct {
	DiscountActivity bool `yaml:"discount_activity"`
	AdvActivity      bool `yaml:"adv_activity"`
	LuckDrawActivity bool `yaml:"luck_draw_activity"`
	ChargeActivity   bool `yaml:"charge_activity"`
}

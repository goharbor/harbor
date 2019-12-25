package driver

import (
	"time"

	bc "github.com/astaxie/beego/cache"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/log"
)

const (
	harborCfg = "harborCfg"
)

// CachedDriver to cache configure properties,
// disable cache by setting cfg_cache_interval_seconds=0
type CachedDriver struct {
	cache bc.Cache
	// Interval - the interval to refresh cache, no cache when Interval=0
	Interval time.Duration
	Driver   Driver
}

// Init - Initial the cache before usage
func (c *CachedDriver) Init() {
	c.cache = bc.NewMemoryCache()
}

func (c *CachedDriver) refresh() error {
	log.Debug("Refresh configure properties from store.")
	cfg, err := c.Driver.Load()
	if err != nil {
		log.Errorf("Failed to load config %+v", err)
		return err
	}
	c.updateInterval(cfg)
	if c.Interval == 0 {
		return nil
	}
	if err = c.cache.Put(harborCfg, cfg, c.Interval); err != nil {
		log.Errorf("Failed to save to cache %v", err)
	}
	return nil
}

func (c *CachedDriver) updateInterval(cfg map[string]interface{}) {
	i, exist := cfg[common.CfgCacheIntervalSeconds]
	if !exist {
		return
	}
	secInt, ok := i.(int)
	if !ok || secInt < 0 {
		return
	}
	c.Interval = time.Duration(secInt) * time.Second
}

// Load - load config item
func (c *CachedDriver) Load() (map[string]interface{}, error) {
	// No cache when Interval = 0
	if c.Interval == 0 {
		cfg, err := c.Driver.Load()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if !c.cache.IsExist(harborCfg) {
		err := c.refresh()
		if err != nil {
			return nil, err
		}
	}
	var cfgMap map[string]interface{}
	obj := c.cache.Get(harborCfg)
	cfgMap, ok := obj.(map[string]interface{})
	if !ok {
		log.Errorf("Failed to retrieve map[string]interface{} from cache, object type is %#v", obj)
	}
	return cfgMap, nil
}

// Save - save config item into config driver
func (c *CachedDriver) Save(cfg map[string]interface{}) error {
	if c.Interval > 0 {
		err := c.cache.Put(harborCfg, cfg, c.Interval)
		if err != nil {
			log.Errorf("Failed to refresh cache %v", err)
		}
	}
	return c.Driver.Save(cfg)
}
